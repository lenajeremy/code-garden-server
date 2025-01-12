package api

import (
	"code-garden-server/config"
	"code-garden-server/utils"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	Handler     func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool)
	PreHandler  func(path string)
	PostHandler func(w http.ResponseWriter, r *http.Request)
}

func NewLoggerMiddleware() Middleware {
	preHandler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		return w, r, true
	}

	postHandler := func(w http.ResponseWriter, r *http.Request) {
		status := w.Header().Get("Status")
		path := r.URL.String()
		log.Printf("%s \t %s\n", path, status)
	}

	m := Middleware{PostHandler: postHandler, Handler: preHandler}
	return m
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

	return Middleware{Handler: handler, PreHandler: preHandler}
}

func NewAuthMiddleware(s *Server) Middleware {
	handler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Invalid Token", Error: "invalid token"})
			return w, r, false
		}
		token := authHeader[len("Bearer "):]
		_, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
			key := config.GetEnv("JWT_SECRET")
			return []byte(key), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Expired", Error: err.Error()})
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Malformed", Error: err.Error()})
			} else {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized!", Error: err.Error()})
			}
			return w, r, false
		}

		// TODO: Find a way to get the user details and store it in the request context
		return w, r, true
	}

	return Middleware{
		Handler: handler,
	}
}

func setCorsHeaders(w http.ResponseWriter, isOptions bool) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if isOptions {
		w.WriteHeader(http.StatusNoContent)
	}
}
