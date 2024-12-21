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
	Error  error  `json:"error"`
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

		time.Sleep(time.Second * 5)

		defer func(body io.ReadCloser) {
			err := body.Close()
			if err != nil {
				return
			}
		}(r.Body)

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request")})
			return
		}

		err = json.Unmarshal(bytes, &body)
		if err != nil {
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request: failed to parse req body")})
			return
		}

		var filename, command string

		if body.Language == "javascript" {
			command = "node"
			filename = "main.js"
		} else if body.Language == "python" {
			command = "python3"
			filename = "main.py"
		} else if body.Language == "typescript" {
			command = "ts-node"
			filename = "main.ts"
		} else {
			writeRes(w, response{"", http.StatusBadRequest, errors.New("bad request: unsupported language")})
			return
		}

		file, err := os.Create(filename)
		if err != nil {
			writeRes(w, response{"", http.StatusInternalServerError, errors.New("internal server error: failed to create file")})
			return
		}

		_, err = file.WriteString(body.Code)
		if err != nil {
			writeRes(w, response{"", http.StatusInternalServerError, errors.New("internal server error: failed to write to file")})
			return
		}

		cmd := exec.Command(command, filename)

		output, err := cmd.CombinedOutput()
		if err != nil {
			writeRes(w, response{string(output), http.StatusInternalServerError, errors.New("failed to run command")})
			return
		}

		writeRes(w, response{string(output), http.StatusOK, nil})
		return
	})

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
