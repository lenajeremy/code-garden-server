package api

import (
	"fmt"
	"net/http"
)

type Middleware struct {
	Handler    func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool)
	PreHandler *func(path string)
}


func NewCorsMiddleware(s *Server) Middleware {

	registeredPaths := map[string]bool{}

	// sets up options handler for every request
	preHandler := func(path string) {
		if _, ok := registeredPaths[path]; !ok {
			s.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodOptions, path), func(w http.ResponseWriter, r *http.Request) {
				setCorsHeaders(w, true)
			})
		}
		registeredPaths[path] = true
	}

	// handles every non-option http method
	handler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		setCorsHeaders(w, false)
		return w, r, true
	}

	return Middleware{Handler: handler, PreHandler: &preHandler}
}

func setCorsHeaders(w http.ResponseWriter, isOptions bool) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if isOptions {
		w.WriteHeader(http.StatusNoContent)
	}
}
