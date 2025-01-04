package server

import (
	c "code-garden-server/controllers"
	"code-garden-server/routes"
)

func InitServer(port int) {
	s := New(port)

	appRoutes := []routes.Route{
		{"POST", "/", c.HandleRunCode},
		{"GET", "/hello", c.HelloController},
	}

	for _, r := range appRoutes {
		s.AddRoute(r)
	}

	s.Start()
}
