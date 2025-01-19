package api

import (
	"code-garden-server/internal/api/handlers"
	"code-garden-server/internal/database"
	"github.com/docker/docker/client"
	"net/http"
	"time"
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
	authMiddleware := NewAuthMiddleware(s)

	defaultRouter := s.DefaultRouter()

	defaultRouter.Use(corsMiddleware)
	defaultRouter.Use(loggerMiddleware)
	defaultRouter.Use(delayMiddleware)

	defaultRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	// main app routes
	appRouter := defaultRouter.Group("/")
	appRouter.Use(authMiddleware)

	appRouter.Post("/run-unsafe", codeHandler.RunCodeUnsafe)
	appRouter.Get("/hello", codeHandler.SayHello)
	appRouter.Get("/containers", dockerHandler.ListContainers)
	appRouter.Post("/run-safe", dockerHandler.RunCodeSafe)

	// snippets sharing and retrieving
	appRouter.Post("/snippet/create", codeHandler.CreateCodeSnippet)
	appRouter.Get("/snippet/{publicId}", codeHandler.GetSnippet)
	appRouter.Put("/snippet/{publicId}", codeHandler.UpdateSnippet)

	appRouter.Get("/snippets/mine", codeHandler.GetUserSnippets)

	// authentication router
	auth := defaultRouter.Group("auth")
	auth.Post("/login-with-email", authHandler.LoginWithEmail)
	auth.Post("/login-with-password", authHandler.LoginWithPassword)

	auth.Post("/register-with-email", authHandler.RegisterWithEmail)
	auth.Post("/register-with-password", authHandler.RegisterWithPassword)

	auth.Post("/verify-email/{token}", authHandler.VerifyUserEmail)
	auth.Post("/sign-in-with-token/{token}", authHandler.SignInWithToken)

	s.Start()
}
