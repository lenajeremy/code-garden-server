package handlers

import (
	"code-garden-server/internal/services/docker"
	"code-garden-server/utils"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"net/http"
)

type DockerHandler struct {
	service *docker.Service
}

func NewDockerHandler(dc *client.Client) *DockerHandler {
	dockerService := docker.NewDockerService(dc)
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

func (d *DockerHandler) RunCodeSafe(w http.ResponseWriter, r *http.Request) {
	type reqbody struct {
		Code     string `json:"code"`
		Language string `json:"language"`
	}

	var body reqbody

	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			return
		}
	}(r.Body)

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: %s", err.Error())})
		return
	}

	err = json.Unmarshal(bytes, &body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: failed to parse req body, %s", err.Error())})
		return
	}

	res, err := d.service.RunLanguageContainer(docker.Language(body.Language), body.Code)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Error: fmt.Sprintf("internal server error: %s", err.Error())})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Output: res})
}
