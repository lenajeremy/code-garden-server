package api

import (
	"code-garden-server/internal/api/handlers"
	"code-garden-server/internal/database"
	"math/rand"
	"net/http"
	"time"

	"github.com/docker/docker/client"
)

func InitServer(p int, dc *client.Client, dbc *database.DBClient) {
	s := NewServer(p, dc, dbc)

	codeHandler := handlers.NewCodeHandler(dbc)
	dockerHandler := handlers.NewDockerHandler(dc, dbc)

	delayMiddleware := Middleware{func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		delay := time.Second * time.Duration(1+rand.Intn(4))
		time.Sleep(delay)
		return w, r, true
	}, nil}

	corsMiddleware := NewCorsMiddleware(s)

	r := s.DefaultRouter()
	r.Use(corsMiddleware)
	r.Use(delayMiddleware)

	r.Post("/run-unsafe", codeHandler.RunCodeUnsafe)
	r.Get("/hello", codeHandler.SayHello)
	r.Get("/containers", dockerHandler.ListContainers)
	r.Post("/run-safe", dockerHandler.RunCodeSafe)

	// snippets sharing and retrieving
	r.Post("/snippet/create", codeHandler.CreateCodeSnippet)
	r.Get("/snippet/{publicId}", codeHandler.GetSnippet)
	r.Put("/snippet/{publicId}", codeHandler.UpdateSnippet)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	s.Start()
}
