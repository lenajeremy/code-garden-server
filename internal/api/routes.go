package api

import (
	"code-garden-server/internal/api/handlers"
	"code-garden-server/internal/database"
	"net/http"

	"github.com/docker/docker/client"
)

func InitServer(p int, dc *client.Client, dbc *database.DBClient) {
	s := NewServer(p, dc, dbc)

	codeHandler := handlers.NewCodeHandler(dbc)
	dockerHandler := handlers.NewDockerHandler(dc, dbc)

	corsMiddleware := Middleware{func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return w, r, false
		}

		return w, r, true
	}}

	r := s.DefaultRouter()
	r.Use(corsMiddleware)

	r.Post("/run-unsafe", codeHandler.RunCodeUnsafe)
	r.Get("/hello", codeHandler.SayHello)
	r.Get("/containers", dockerHandler.ListContainers)
	r.Post("/run-safe", dockerHandler.RunCodeSafe)

	// snippets sharing and retrieving
	r.Post("/share", codeHandler.ShareCodeSnippet)
	r.Get("/get-snippet/{publicId}", codeHandler.GetSnippet)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	s.Start()
}
