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
	// var redirectToHttps = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	// })

	// go func() {
	// 	log.Fatal(http.ListenAndServe(":80", redirectToHttps))
	// }()

	// log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", s.Port),
	// "/etc/letsencrypt/live/cgs.craftmycv.xyz/fullchain.pem",
	// "/etc/letsencrypt/live/cgs.craftmycv.xyz/privkey.pem",
	// s.mux,
	// ))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux))
}
func (s *Server) DefaultRouter() *Router {
	return newRouter(s.mux, "/")
}

type excludeMiddlewareSet = map[*Middleware]bool

func newExcludeMiddlewareSet(ms ...*Middleware) excludeMiddlewareSet {
	s := map[*Middleware]bool{}
	for _, m := range ms {
		s[m] = true
	}
	return s
}

type Router struct {
	mux           *http.ServeMux
	middlewares   []*Middleware
	path          string
	middlewareSet excludeMiddlewareSet
}

func newRouter(mux *http.ServeMux, path string) *Router {
	return &Router{mux: mux, path: path, middlewareSet: map[*Middleware]bool{}}
}

func (r *Router) Group(path string, middlewares ...*Middleware) *Router {
	path = filepath.Join(r.path, path)
	rout := newRouter(r.mux, path)
	rout.middlewares = append(r.middlewares, middlewares...)

	for _, m := range rout.middlewares {
		r.middlewareSet[m] = true
	}

	return rout
}

func (r *Router) Use(m *Middleware) {
	r.middlewares = append(r.middlewares, m)
	r.middlewareSet[m] = true
}

func (r *Router) Get(path string, handler http.HandlerFunc, excludeMiddleware ...*Middleware) {
	resolvedPath := filepath.Join(r.path, path)
	emSet := newExcludeMiddlewareSet(excludeMiddleware...)

	r.RunAllPreMiddlewareHandlers(resolvedPath, emSet)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodGet, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddlewareHandlers(w, req, true, emSet)
		if !ok {
			return
		} else {
			handler(w, req)
			r.RunAllPostMiddlewareHandlers(w, req, emSet)
		}
	})
}

func (r *Router) Post(path string, handler http.HandlerFunc, excludeMiddleware ...*Middleware) {
	resolvedPath := filepath.Join(r.path, path)
	emSet := newExcludeMiddlewareSet(excludeMiddleware...)

	r.RunAllPreMiddlewareHandlers(resolvedPath, emSet)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPost, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddlewareHandlers(w, req, true, emSet)
		if !ok {
			return
		} else {
			handler(w, req)
			r.RunAllPostMiddlewareHandlers(w, req, emSet)
		}
	})
}

func (r *Router) Put(path string, handler http.HandlerFunc, excludeMiddleware ...*Middleware) {
	resolvedPath := filepath.Join(r.path, path)
	emSet := newExcludeMiddlewareSet(excludeMiddleware...)

	r.RunAllPreMiddlewareHandlers(resolvedPath, emSet)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPut, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddlewareHandlers(w, req, true, emSet)
		if !ok {
			return
		} else {
			handler(w, req)
			r.RunAllPostMiddlewareHandlers(w, req, emSet)
		}
	})
}

func (r *Router) Delete(path string, handler http.HandlerFunc, excludeMiddleware ...*Middleware) {
	resolvedPath := filepath.Join(r.path, path)
	emSet := newExcludeMiddlewareSet(excludeMiddleware...)

	r.RunAllPreMiddlewareHandlers(resolvedPath, emSet)
	r.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodDelete, resolvedPath), func(w http.ResponseWriter, req *http.Request) {
		w, req, ok := r.RunAllMiddlewareHandlers(w, req, true, emSet)
		if !ok {
			return
		}
		handler(w, req)
		r.RunAllPostMiddlewareHandlers(w, req, emSet)
	})
}

func (r *Router) RunAllMiddlewareHandlers(w http.ResponseWriter, req *http.Request, ok bool, excluding excludeMiddlewareSet) (http.ResponseWriter, *http.Request, bool) {
	for _, m := range r.middlewares {
		if m.Handler == nil {
			break
		}

		if excluding[m] {
			continue
		}

		w, req, ok = m.Handler(w, req)
		if !ok {
			break
		}
	}
	return w, req, ok
}

func (r *Router) RunAllPreMiddlewareHandlers(path string, excluding excludeMiddlewareSet) {
	for _, m := range r.middlewares {
		if !excluding[m] && m.PreHandler != nil {
			m.PreHandler(path)
		}
	}
}

func (r *Router) RunAllPostMiddlewareHandlers(w http.ResponseWriter, req *http.Request, excluding excludeMiddlewareSet) {
	for _, m := range r.middlewares {
		if !excluding[m] && m.PostHandler != nil {
			m.PostHandler(w, req)
		}
	}
}
