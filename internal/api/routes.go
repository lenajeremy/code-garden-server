package api

import (
	"code-garden-server/internal/api/handlers"
	"code-garden-server/internal/database"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"
)

func InitServer(p int, dc *client.Client, dbc *database.DBClient, rds *redis.Client) {
	s := NewServer(p, dc, dbc, rds)

	codeHandler := handlers.NewCodeHandler(dbc, rds)
	dockerHandler := handlers.NewDockerHandler(dc, dbc)
	authHandler := handlers.NewAuthHandler(dbc, rds)

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

	defaultRouter.Use(&corsMiddleware)
	defaultRouter.Use(&loggerMiddleware)
	defaultRouter.Use(&delayMiddleware)

	defaultRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	// main app routes
	appRouter := defaultRouter.Group("/")
	appRouter.Use(&authMiddleware)

	// appRouter.Post("/run-unsafe", codeHandler.RunCodeUnsafe)
	appRouter.Get("/hello", codeHandler.SayHello)
	appRouter.Get("/containers", dockerHandler.ListContainers)
	appRouter.Post("/code-runner", dockerHandler.RunCodeSafe)
	appRouter.Post("/code-runner/no-auth", dockerHandler.RunCodeSafeNoAuth, &authMiddleware)

	// snippets sharing and retrieving
	appRouter.Post("/snippet/create", codeHandler.CreateCodeSnippet)
	appRouter.Get("/snippet/{publicId}", codeHandler.GetSnippet)
	appRouter.Get("/snippet/{publicId}/no-auth", codeHandler.GetSnippetNoAuth, &authMiddleware)
	appRouter.Put("/snippet/{publicId}", codeHandler.UpdateSnippet)
	appRouter.Delete("/snippet/{publicId}", codeHandler.DeleteSnippet)
	appRouter.Post("/snippet/{publicId}/fork", codeHandler.ForkSnippet)

	appRouter.Get("/snippets/mine", codeHandler.GetUserSnippets)

	// authentication router
	auth := defaultRouter.Group("auth")
	auth.Post("/login-with-email", authHandler.LoginWithEmail)
	auth.Post("/login-with-password", authHandler.LoginWithPassword)

	auth.Post("/register-with-email", authHandler.RegisterWithEmail)
	auth.Post("/register-with-password", authHandler.RegisterWithPassword)

	auth.Post("/verify-email/{token}", authHandler.VerifyUserEmail)
	auth.Post("/sign-in-with-token/{token}", authHandler.SignInWithToken)
	auth.Post("/request-password-reset", authHandler.RequestPasswordReset)
	auth.Post("/reset-password", authHandler.ResetPassword)

	s.Start()
}
