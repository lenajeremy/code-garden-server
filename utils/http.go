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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Status", strconv.Itoa(rb.Status))

	if rb.Status >= 200 && rb.Status < 300 {
		w.Header().Set("Message", rb.Message)
	} else {
		w.Header().Set("Message", rb.Error)
	}

	err := json.NewEncoder(w).Encode(rb)
	if err != nil {
		http.Error(w, "internal server error: failed to encode response", http.StatusInternalServerError)
		return
	}
}
