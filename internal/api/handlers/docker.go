package handlers

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/auth"
	"code-garden-server/internal/services/docker"
	"code-garden-server/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

var userToAllowedRunCount = map[string]int{}
var userToNextResetTime = map[string]time.Time{}

const REQUESTS_ALLOWED_PER_MINUTE = 20
const REQUESTS_ALLOWED_PER_MINUTE_NO_AUTH = 5

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

	u := auth.AuthUser(r)
	uid := u.ID.String()

	if _, okay := userToNextResetTime[uid]; !okay {
		userToAllowedRunCount[uid] = REQUESTS_ALLOWED_PER_MINUTE
		userToNextResetTime[uid] = time.Now().Add(time.Minute)
	}

	nextResetTime := userToNextResetTime[uid]
	fmt.Println(userToNextResetTime, userToAllowedRunCount)

	if time.Now().After(nextResetTime) {
		userToAllowedRunCount[uid] = REQUESTS_ALLOWED_PER_MINUTE
	}

	if userToAllowedRunCount[uid] == 0 {
		utils.WriteRes(w, utils.Response{Status: http.StatusTooManyRequests, Message: "Too many requests have been sent. Wait for a minute", Error: "Limit exceeded"})
		return
	}

	defer func() {
		userToAllowedRunCount[uid] -= 1
		if time.Now().After(nextResetTime) {
			userToNextResetTime[uid] = time.Now().Add(time.Minute)
		}
		fmt.Println(userToNextResetTime, userToAllowedRunCount)
	}()

	res, err := d.service.RunLanguageContainer(docker.Language(body.Language), body.Code)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Data: res, Error: fmt.Sprintf("internal server error: %s", err.Error()), Message: "Error"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Data: res, Message: "Success"})
}

func (d *DockerHandler) RunCodeSafeNoAuth(w http.ResponseWriter, r *http.Request) {
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

	uid := r.RemoteAddr

	if _, okay := userToNextResetTime[uid]; !okay {
		userToAllowedRunCount[uid] = REQUESTS_ALLOWED_PER_MINUTE_NO_AUTH
		userToNextResetTime[uid] = time.Now().Add(time.Minute)
	}

	nextResetTime := userToNextResetTime[uid]

	if time.Now().After(nextResetTime) {
		userToAllowedRunCount[uid] = REQUESTS_ALLOWED_PER_MINUTE_NO_AUTH
	}

	if userToAllowedRunCount[uid] == 0 {
		utils.WriteRes(w, utils.Response{Status: http.StatusTooManyRequests, Message: "Too many requests have been sent. Wait for a minute", Error: "Limit exceeded"})
		return
	}

	defer func() {
		userToAllowedRunCount[uid] -= 1
		if time.Now().After(nextResetTime) {
			userToNextResetTime[uid] = time.Now().Add(time.Minute)
		}
		fmt.Println(userToNextResetTime, userToAllowedRunCount)
	}()

	res, err := d.service.RunLanguageContainer(docker.Language(body.Language), body.Code)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Data: res, Error: fmt.Sprintf("internal server error: %s", err.Error()), Message: "Error"})
		return
	}

	utils.WriteRes(w, utils.Response{Status: http.StatusOK, Data: res, Message: "Success"})
}
