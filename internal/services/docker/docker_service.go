package docker

import (
	"archive/tar"
	"bytes"
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
	client *client.Client
}

func NewDockerService(dc *client.Client) *Service {
	s := &Service{dc}
	err := s.SetupClient()
	if err != nil {
		log.Fatal(err)
	}
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
	containers, err := ds.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return []types.Container{}, err
	}

	return containers, nil
}

func (ds *Service) RunLanguageContainer(lang Language, codeSrc string) (string, error) {
	// Create a context with timeout to prevent hanging containers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	resp, err := ds.client.ContainerCreate(ctx, config, &container.HostConfig{
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

		if err := ds.client.ContainerRemove(removeCtx, resp.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			log.Printf("failed to remove container %s: %v", resp.ID, err)
		} else {
			log.Printf("container %s removed", resp.ID)
		}
	}()

	if err := ds.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	log.Printf("Container %s started in %v", resp.ID, time.Since(start))

	// Attach to container with improved error handling
	attachResp, err := ds.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Write code to container in a separate goroutine
	go func() {
		defer attachResp.CloseWrite()
		if _, err := io.Copy(attachResp.Conn, strings.NewReader(codeSrc)); err != nil {
			log.Printf("error writing to container: %v", err)
		}
	}()

	// Create a buffer to store container output
	var outputBuf bytes.Buffer
	outputDone := make(chan error)

	// Read container output in separate goroutine
	go func() {
		_, err := stdcopy.StdCopy(&outputBuf, &outputBuf, attachResp.Reader)
		if err != nil {
			log.Printf("Error reading container output: %v", err)
			outputDone <- err
			return
		}
		outputDone <- nil
	}()

	// Wait for container completion with timeout
	statusCh, errCh := ds.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return "", fmt.Errorf("container wait error: %w", err)
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return "", fmt.Errorf("container exited with status %d: %s",
				status.StatusCode, status.Error.Message)
		}
	case <-ctx.Done():
		err = fmt.Errorf("container execution timed out after %v", time.Since(start))
		return "", fmt.Errorf("container execution timed out after %v", time.Since(start))
	}

	// Wait for output processing to complete
	if err := <-outputDone; err != nil {
		return "", fmt.Errorf("error reading container output: %w", err)
	}

	log.Printf("Container %s completed execution in %v", resp.ID, time.Since(start))

	fmt.Printf("OUTPUT\n\n"+
		"%s\n", outputBuf.String())

	return outputBuf.String(), nil
}

func (ds *Service) BuildLanguageImage(language Language) error {
	log.Println("building image", language)
	dockerfile := LanguageToDockerFileMap[language]

	buildContext, err := createBuildContext(dockerfile, "run.sh")
	if err != nil {
		return err
	}

	imgBuildResponse, err := ds.client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
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
