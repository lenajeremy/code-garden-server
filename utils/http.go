package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Error   string      `json:"error"`
}

func WriteRes(w http.ResponseWriter, rb Response) {
	w.WriteHeader(rb.Status)
	w.Header().Set("Content-Type", "text/json")
	w.Header().Set("Status", strconv.Itoa(rb.Status))

	err := json.NewEncoder(w).Encode(rb)
	if err != nil {
		http.Error(w, "internal server error: failed to encode response", http.StatusInternalServerError)
		return
	}
}
