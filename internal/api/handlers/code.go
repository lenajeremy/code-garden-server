package handlers

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/database/models"
	"code-garden-server/internal/services/docker"
	"code-garden-server/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type CodeHandler struct {
	DbClient *database.DBClient
}

type reqbody struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

func NewCodeHandler(dbClient *database.DBClient) *CodeHandler {
	return &CodeHandler{
		dbClient,
	}
}

func (*CodeHandler) SayHello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	_, err := fmt.Fprintf(w, "Hello %s!", name)
	if err != nil {
		return
	}
}

func (*CodeHandler) RunCodeUnsafe(w http.ResponseWriter, r *http.Request) {
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

	file, err := os.CreateTemp("/Users/jeremiahlena/Desktop", filename)
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

	commands = append(commands, file.Name())
	cmd := exec.Command(commands[0], commands[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.WriteRes(w, utils.Response{Data: string(output), Status: http.StatusInternalServerError, Error: fmt.Sprintf("failed to run command %s", err), Message: "Error"})
		return
	}

	utils.WriteRes(w, utils.Response{Data: string(output), Status: http.StatusOK, Message: "Success"})
}

func (c *CodeHandler) ShareCodeSnippet(w http.ResponseWriter, r *http.Request) {

	type sharecodereqbody struct {
		Code     string `json:"code"`
		Language string `json:"language"`
		Output   string `json:"output"`
	}

	var body sharecodereqbody

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(r.Body)

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&body)

	if _, ok := docker.LanguageToImageMap[docker.Language(body.Language)]; !ok {
		utils.WriteRes(w, utils.Response{Data: nil, Message: "Unsupported Language", Status: http.StatusBadRequest, Error: "unsupported language"})
		return
	}

	snippet := models.Snippet{Code: body.Code, Language: body.Language, Output: body.Output}
	tx := c.DbClient.Create(&snippet)
	if tx.Error != nil {
		utils.WriteRes(w, utils.Response{Data: nil, Message: "Failed to create snipped", Status: http.StatusInternalServerError, Error: tx.Error.Error()})
		return
	}

	utils.WriteRes(w, utils.Response{Data: snippet, Message: "Snippet created successfully", Status: http.StatusCreated, Error: ""})
}
