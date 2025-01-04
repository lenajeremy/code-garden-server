package main

import (
	"code-garden-server/server"
	"log"
)

func main() {
	PORT := 8000
	log.Print("starting server on port 8000")
	server.InitServer(PORT)
}
