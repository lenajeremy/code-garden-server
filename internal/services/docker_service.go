package services

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func ListRunningContainers(client *client.Client) ([]types.Container, error) {
	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return []types.Container{}, err
	}

	return containers, nil
}

func NewDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.45"), client.FromEnv)
	if err != nil {
		return nil, err
	}

	return cli, nil
}
