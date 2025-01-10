package api

import (
	"code-garden-server/internal/database"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/docker/docker/client"
)

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

type Server struct {
	routes       []Route
	Port         int
	mux          *http.ServeMux
	dockerClient *client.Client
	db           *database.DBClient
}

// NewServer creates a new server with the default values
func NewServer(port int, dockerClient *client.Client, db *database.DBClient) *Server {
	mux := http.NewServeMux()
	return &Server{
		[]Route{},
		port,
		mux,
		dockerClient,
		db,
	}
}

// Start sets up all the routes and starts the server
func (s *Server) Start() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux))
}

func (s *Server) DefaultRouter() *Router {
	return newRouter(s.mux, "/")
}

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
	path        string
}

func newRouter(mux *http.ServeMux, path string) *Router {
	return &Router{mux: mux, path: path}
}

func (r *Router) Group(path string, middlewares ...Middleware) *Router {
	path = filepath.Join(r.path, path)
	rout := newRouter(r.mux, path)
	rout.middlewares = middlewares
	return rout
}

func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}


func (r *Router) Get(path string, handler http.HandlerFunc) {
	resolvedPath := filepath.Join(r.path, path)

	r.RunAllPreMiddleware(resolvedPath)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodGet, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddleware(w, req, true)
		if !ok {
			return
		} else {
			handler(w, req)
		}
	})
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	resolvedPath := filepath.Join(r.path, path)

	r.RunAllPreMiddleware(resolvedPath)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPost, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddleware(w, req, true)
		if !ok {
			return
		} else {
			handler(w, req)
		}
	})
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	resolvedPath := filepath.Join(r.path, path)

	r.RunAllPreMiddleware(resolvedPath)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPut, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddleware(w, req, true)
		if !ok {
			return
		} else {
			handler(w, req)
		}
	})
}

func (r *Router) RunAllMiddleware(w http.ResponseWriter, req *http.Request, ok bool) (http.ResponseWriter, *http.Request, bool) {
	for _, m := range r.middlewares {
		w, req, ok = m.Handler(w, req)
		if !ok {
			break
		}
	}
	return w, req, ok
}

func (r *Router) RunAllPreMiddleware(path string) {
	for _, m := range r.middlewares {
		if m.PreHandler != nil {
			f := *m.PreHandler
			f(path)
		}
	}
}
