package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type reqbody struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type response struct {
	Output string `json:"output"`
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func writeRes(w http.ResponseWriter, rb response) {
	w.WriteHeader(rb.Status)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(rb)
	if err != nil {
		http.Error(w, "internal server error: failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body reqbody

		time.Sleep(-5 * time.Second)

		defer func(body io.ReadCloser) {
			err := body.Close()
			if err != nil {
				return
			}
		}(r.Body)

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request").Error()})
			return
		}

		err = json.Unmarshal(bytes, &body)
		if err != nil {
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request: failed to parse req body").Error()})
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
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request: unsupported language").Error()})
			return
		}

		file, err := os.Create(filename)
		if err != nil {
			writeRes(w, response{"", http.StatusInternalServerError, errors.New("internal server error: failed to create file").Error()})
			return
		}

		_, err = file.WriteString(body.Code)
		if err != nil {
			writeRes(w, response{"", http.StatusInternalServerError, errors.New("internal server error: failed to write to file").Error()})
			return
		}

		commands = append(commands, filename)
		cmd := exec.Command(commands[0], commands[1:]...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			writeRes(w, response{string(output), http.StatusInternalServerError, "failed to run command"})
			return
		}

		writeRes(w, response{string(output), http.StatusOK, ""})
		return
	})

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
