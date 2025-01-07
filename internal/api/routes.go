package api

import (
	"github.com/docker/docker/client"
	"net/http"
)

func InitServer(p int, dc *client.Client) {
	s := NewServer(p, dc, nil)

	/* PREVIOUS CODE */

	//codeHandler := handlers.NewCodeHandler()
	//dockerHandler := handlers.NewDockerHandler(dc)

	//appRoutes := []Route{
	//	{"POST", "/run-unsafe", codeHandler.RunCodeUnsafe},
	//	{"GET", "/hello", codeHandler.SayHello},
	//	{"GET", "/containers", dockerHandler.ListContainers},
	//	{"POST", "/run-safe", dockerHandler.RunCodeSafe},
	//}

	evonMedicsHeaderMiddleware := Middleware{func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("EvonMedics", "BABYYYYYYYYYY")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return w, r, true
	}}

	r := s.DefaultRouter()
	v1 := r.Group("/v1", evonMedicsHeaderMiddleware)
	v1.Get("/sing", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("singing for evonmedics"))
	})
	v1.Get("/dance", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("dancing for evonmedics"))
	})
	s.Start()
}
