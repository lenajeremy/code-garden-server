package handlers

import (
	"code-garden-server/internal/database"
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

func NewDockerHandler(dc *client.Client, dbc *database.DBClient) *DockerHandler {
	dockerService := docker.NewDockerService(dc, dbc)
	return &DockerHandler{dockerService}
}

func (d *DockerHandler) ListContainers(w http.ResponseWriter, _ *http.Request) {
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
	type reqBody struct {
		Code     string `json:"code"`
		Language string `json:"language"`
	}

	var body reqBody

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(r.Body)

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: failed to parse req body, %s", err.Error())})
		return
	}

	// bytes, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: %s", err.Error())})
	// 	return
	// }

	// err = json.Unmarshal(bytes, &body)
	// if err != nil {
	// 	utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: failed to parse req body, %s", err.Error())})
	// 	return
	// }

	res, err := d.service.RunLanguageContainer(docker.Language(body.Language), body.Code)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Data: res, Error: fmt.Sprintf("internal server error: %s", err.Error()), Message: "Error"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Data: res, Message: "Success"})
}
