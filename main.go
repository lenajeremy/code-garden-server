package main

import (
	"code-garden-server/internal/api"
	"code-garden-server/internal/services"
	"log"
)

func main() {
	dckClient, err := services.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	PORT := 3000
	log.Printf("starting server on port %d", PORT)
	api.InitServer(PORT, dckClient)
}
