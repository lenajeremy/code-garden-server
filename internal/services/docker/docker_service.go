package docker

import (
	"archive/tar"
	"bytes"
	"code-garden-server/internal/database"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Service struct {
	dockerClient   *client.Client
	databaseClient *database.DBClient
}

func NewDockerService(dc *client.Client, dbClient *database.DBClient) *Service {
	s := &Service{dc, dbClient}
	return s
}

func NewDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.45"), client.FromEnv)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (ds *Service) ListRunningContainers() ([]types.Container, error) {
	containers, err := ds.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return []types.Container{}, err
	}

	return containers, nil
}

func (ds *Service) RunLanguageContainer(lang Language, codeSrc string) (string, error) {
	// Create a context with timeout to prevent hanging containers
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	image, ok := LanguageToImageMap[lang]
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", lang)
	}

	// Create container config with improved settings
	config := &container.Config{
		Image:        image,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		StdinOnce:    true,
	}

	start := time.Now()
	resp, err := ds.dockerClient.ContainerCreate(ctx, config, &container.HostConfig{
		CapDrop: []string{"ALL"},
	}, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure container is removed even if we error out
	defer func() {
		timeout := 5 * time.Second
		removeCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := ds.dockerClient.ContainerRemove(removeCtx, resp.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			log.Printf("failed to remove container %s: %v", resp.ID, err)
		} else {
			log.Printf("container %s removed", resp.ID)
		}
	}()

	if err := ds.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	log.Printf("Container %s started in %v", resp.ID, time.Since(start))

	// Attach to container with improved error handling
	attachResp, err := ds.dockerClient.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Create error channels for input/output operations
	inputDone := make(chan error, 1)
	outputDone := make(chan error, 1)
	var outputBuf bytes.Buffer

	// Write code to container in a separate goroutine
	go func() {
		defer attachResp.CloseWrite()
		_, err := io.Copy(attachResp.Conn, strings.NewReader(codeSrc))
		inputDone <- err
	}()

	// Read container output in separate goroutine
	go func() {
		_, err := stdcopy.StdCopy(&outputBuf, &outputBuf, attachResp.Reader)
		outputDone <- err
	}()

	// Wait for container completion with timeout
	statusCh, errCh := ds.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return "", fmt.Errorf("container wait error: %w", err)
	case status := <-statusCh:
		// Handle input completion
		if err := <-inputDone; err != nil {
			return "", fmt.Errorf("error writing to container: %w", err)
		}

		// Handle output completion
		if err := <-outputDone; err != nil {
			return "", fmt.Errorf("error reading container output: %w", err)
		}

		log.Printf("Container %s completed execution in %v", resp.ID, time.Since(start))

		if status.StatusCode != 0 {
			// Get container logs for error details
			logOptions := container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Since:      start.Format(time.RFC3339),
			}

			logs, err := ds.dockerClient.ContainerLogs(ctx, resp.ID, logOptions)
			if err != nil {
				return "", fmt.Errorf("container failed with status %d and error getting logs: %w",
					status.StatusCode, err)
			}
			defer logs.Close()

			var errorOutput bytes.Buffer
			if _, err := stdcopy.StdCopy(&errorOutput, &errorOutput, logs); err != nil {
				return "", fmt.Errorf("container failed with status %d and error reading logs: %w",
					status.StatusCode, err)
			}

			// return 10kb of error output
			return string(errorOutput.Bytes()[0 : 1024*10]), fmt.Errorf("container failed with status %d",
				status.StatusCode)
		}

	case <-ctx.Done():
		return "", fmt.Errorf("container execution timed out after %v", time.Since(start))
	}

	output := outputBuf.String()
	log.Printf("Output:\n%s", output)
	return output, nil
}

func (ds *Service) BuildLanguageImage(language Language) error {
	log.Println("building image", language)
	dockerfile := LanguageToDockerFileMap[language]

	buildContext, err := createBuildContext(dockerfile, "run.sh")
	if err != nil {
		return err
	}

	imgBuildResponse, err := ds.dockerClient.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Tags:       []string{LanguageToImageMap[language]},
		Dockerfile: dockerfile,
		Remove:     false,
	})

	if err != nil {
		log.Println("error while building image", err)
		return err
	}

	defer imgBuildResponse.Body.Close()
	body, err := io.ReadAll(imgBuildResponse.Body)
	if err != nil {
		log.Println("error reading build response body:", err)
		return err
	}
	fmt.Printf(`
----------------------------------------
     BUILD OUTPUT FOR "%s":
----------------------------------------

%s
`, language, string(body))

	return err
}

func (ds *Service) SetupClient() error {
	// Build all language images
	for _, language := range SupportedLanguages {
		err := ds.BuildLanguageImage(language)
		if err != nil {
			return err
		}
	}
	return nil
}

func createBuildContext(dockerfile string, files ...string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Add Dockerfile to tar archive
	dockerfileContent, err := os.ReadFile(dockerfile)
	if err != nil {
		return nil, err
	}

	hdr := &tar.Header{
		Name: dockerfile,
		Mode: 0600,
		Size: int64(len(dockerfileContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write(dockerfileContent); err != nil {
		return nil, err
	}

	// Add additional files to tar archive
	for _, file := range files {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		hdr := &tar.Header{
			Name: file,
			Mode: 0600,
			Size: int64(len(fileContent)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(fileContent); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}
