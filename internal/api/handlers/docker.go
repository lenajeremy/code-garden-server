package handlers

import (
	"code-garden-server/internal/services"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"net/http"
)

type DockerHandler struct {
	dockerClient *client.Client
}

func NewDockerHandler(dc *client.Client) *DockerHandler {
	return &DockerHandler{dc}
}

func (d *DockerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := services.ListRunningContainers(d.dockerClient)
	log.Printf("%#v\n", containers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type res struct {
		Containers []types.Container `json:"containers"`
		Err        error             `json:"err"`
	}

	var rr = res{
		Containers: containers,
		Err:        err,
	}

	err = json.NewEncoder(w).Encode(rr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
