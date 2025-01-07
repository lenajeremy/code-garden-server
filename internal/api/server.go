package api

import (
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"net/http"
	"path/filepath"
)

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

type Server struct {
	isRunning    bool
	routes       []Route
	Port         int
	mux          *http.ServeMux
	dockerClient *client.Client
}

// NewServer creates a new server with the default values
func NewServer(port int, dockerClient *client.Client, db interface{}) *Server {
	mux := http.NewServeMux()
	return &Server{
		false,
		[]Route{},
		port,
		mux,
		dockerClient,
	}
}

// Start sets up all the routes and starts the server
func (s *Server) Start() {
	//if s.isRunning {
	//	panic(errors.New("server is already running"))
	//}
	//
	//for _, r := range s.routes {
	//	s.mux.HandleFunc(fmt.Sprintf("%s %s", r.Method, r.Path), r.Handler)
	s.isRunning = true
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux))
}

//// AddRoute adds a new route to the server
//func (s *Server) AddRoute(r Route) {
//	if s.isRunning {
//		panic(errors.New("cannot add a route to a running server"))
//	}
//	s.routes = append(s.routes, r)
//}

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
	rout := newRouter(r.mux, path)
	rout.middlewares = middlewares
	return rout
}

type Middleware struct {
	Handler func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	resolvedPath := filepath.Join(r.path, path)
	fmt.Printf("Resolved path: %s\n", resolvedPath)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", "GET", resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddleware(w, req, true)
		if !ok {
			return
		} else {
			handler(w, req)
		}
	})
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.mux.HandleFunc(fmt.Sprintf("%s %s", "POST", path), handler)
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
