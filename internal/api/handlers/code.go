package handlers

import (
	"code-garden-server/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type CodeHandler struct{}

func NewCodeHandler() *CodeHandler {
	return new(CodeHandler)
}

func (_ *CodeHandler) SayHello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	_, err := fmt.Fprintf(w, "Hello %s!", name)
	if err != nil {
		return
	}
}

func (_ *CodeHandler) RunCodeUnsafe(w http.ResponseWriter, r *http.Request) {
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

	var filename string
	var commands []string

	if body.Language == "javascript" {
		commands = []string{"node"}
		filename = "main.js"
	} else if body.Language == "python" {
		commands = []string{"python3"}
		filename = "main.py"
	} else if body.Language == "typescript" {
		commands = []string{"ts-node"}
		filename = "main.ts"
	} else if body.Language == "go" {
		commands = []string{"go", "run"}
		filename = "run.go"
	} else {
		utils.WriteRes(w, utils.Response{Status: http.StatusBadRequest, Error: fmt.Sprintf("bad request: unsupported language, %s", err)})
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Error: fmt.Sprintf("internal server error: failed to create file, %s", err)})
		return
	}
	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.WriteString(body.Code)
	if err != nil {
		utils.WriteRes(w, utils.Response{Status: http.StatusInternalServerError, Error: fmt.Sprintf("internal server error: failed to write to file, %s", err)})
		return
	}

	commands = append(commands, filename)
	cmd := exec.Command(commands[0], commands[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.WriteRes(w, utils.Response{Output: string(output), Status: http.StatusInternalServerError, Error: fmt.Sprintf("failed to run command %s", err)})
		return
	}

	utils.WriteRes(w, utils.Response{Output: string(output), Status: http.StatusOK})
	return
}
