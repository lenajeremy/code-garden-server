package server

import (
	"code-garden-server/routes"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	isRunning bool
	routes    []routes.Route
	Port      int
	mux       *http.ServeMux
}

// Start sets up all the routes and starts the server
func (s *Server) Start() {
	if s.isRunning {
		panic(errors.New("server is already running"))
	}

	for _, r := range s.routes {
		s.mux.HandleFunc(fmt.Sprintf("%s %s", r.Method, r.Path), r.Handler)
	}

	s.isRunning = true
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux))
}

// AddRoute adds a new route to the server
func (s *Server) AddRoute(r routes.Route) {
	if s.isRunning {
		panic(errors.New("cannot add a route to a running server"))
	}
	s.routes = append(s.routes, r)
}

// New creates a new server with the default values
func New(port int) *Server {
	mux := http.NewServeMux()
	return &Server{
		false,
		[]routes.Route{},
		port,
		mux,
	}
}
