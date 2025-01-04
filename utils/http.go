package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Output string `json:"output"`
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func WriteRes(w http.ResponseWriter, rb Response) {
	w.WriteHeader(rb.Status)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(rb)
	if err != nil {
		http.Error(w, "internal server error: failed to encode response", http.StatusInternalServerError)
		return
	}
}
