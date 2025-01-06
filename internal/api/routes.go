package api

import (
	"code-garden-server/internal/api/handlers"
	"github.com/docker/docker/client"
)

func InitServer(p int, dc *client.Client) {
	s := New(p, dc, nil)

	codeHandler := handlers.NewCodeHandler()
	dockerHandler := handlers.NewDockerHandler(dc)

	appRoutes := []Route{
		{"POST", "/", codeHandler.HandleRunCode},
		{"GET", "/hello", codeHandler.SayHello},
		{"GET", "/containers", dockerHandler.ListContainers},
		{"POST", "/run-safe", dockerHandler.RunCodeSafe},
	}

	for _, r := range appRoutes {
		s.AddRoute(r)
	}

	s.Start()
}
