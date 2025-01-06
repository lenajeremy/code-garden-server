package main

import (
	"code-garden-server/internal/api"
	"code-garden-server/internal/services/docker"
	"log"
)

func main() {
	dckClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}
	defer dckClient.Close()

	PORT := 3000
	log.Printf("starting server on port %d", PORT)
	api.InitServer(PORT, dckClient)
}
