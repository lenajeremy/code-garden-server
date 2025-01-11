package api

import (
	"code-garden-server/internal/api/handlers"
	"code-garden-server/internal/database"
	"net/http"
	"time"

	"github.com/docker/docker/client"
)

func InitServer(p int, dc *client.Client, dbc *database.DBClient) {
	s := NewServer(p, dc, dbc)

	codeHandler := handlers.NewCodeHandler(dbc)
	dockerHandler := handlers.NewDockerHandler(dc, dbc)
	authHandler := handlers.NewAuthHandler(dbc)

	delayMiddleware := Middleware{
		Handler: func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
			time.Sleep(time.Second)
			return w, r, true
		},
	}

	corsMiddleware := NewCorsMiddleware(s)
	loggerMiddleware := NewLoggerMiddleware()

	r := s.DefaultRouter()
	r.Use(corsMiddleware)
	r.Use(delayMiddleware)
	r.Use(loggerMiddleware)

	r.Post("/run-unsafe", codeHandler.RunCodeUnsafe)
	r.Get("/hello", codeHandler.SayHello)
	r.Get("/containers", dockerHandler.ListContainers)
	r.Post("/run-safe", dockerHandler.RunCodeSafe)

	// snippets sharing and retrieving
	r.Post("/snippet/create", codeHandler.CreateCodeSnippet)
	r.Get("/snippet/{publicId}", codeHandler.GetSnippet)
	r.Put("/snippet/{publicId}", codeHandler.UpdateSnippet)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	auth := r.Group("auth")
	auth.Post("/login-with-email", authHandler.LoginWithEmail)

	s.Start()
}
