package handlers

import (
	"code-garden-server/internal/services/docker"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"net/http"
)

type DockerHandler struct {
	service *docker.Service
}

func NewDockerHandler(dc *client.Client) *DockerHandler {
	dockerService := docker.NewDockerService(dc)
	for _, language := range docker.SupportedLanguages {
		if err := dockerService.BuildLanguageImage(language); err != nil {
			log.Fatal(err)
		}
	}
	return &DockerHandler{dockerService}
}

func (d *DockerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := d.service.ListRunningContainers()
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
