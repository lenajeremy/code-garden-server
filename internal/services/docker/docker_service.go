package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
)

type Service struct {
	client *client.Client
}

func NewDockerService(dc *client.Client) *Service {
	return &Service{dc}
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

//func (ds *Service) RunContainer(image string) (string, error) {
//	resp, err := client.ContainerCreate(context.Background(), &container.Config{
//		Image: image,
//	}, nil, nil, nil, "")
//	if err != nil {
//		return "", err
//	}
//
//	if err := client.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
//		return "", err
//	}
//
//	return resp.ID, nil
//}

func (ds *Service) BuildLanguageImage(language Language) error {
	log.Println("building image", language)
	dockerfile := LanguageToDockerFileMap[language]

	buildContext, err := createBuildContext(dockerfile, "go.sh")
	if err != nil {
		return err
	}

	imgBuildResponse, err := ds.client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Tags:       []string{fmt.Sprintf("code-garden-%s", string(language))},
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
	for language, _ := range LanguageToDockerFileMap {
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
