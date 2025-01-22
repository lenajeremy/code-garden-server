package main

import (
	"code-garden-server/internal/api"
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/docker"
	"log"
)

func main() {
	dckClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = dckClient.Close()
	}()

	// create a new database client
	dbClient, err := database.NewDBClient()
	if err != nil {
		log.Fatal("failed to create database client", err)
	}

	// get the client database connection
	conn, err := dbClient.DB.DB()
	if err != nil {
		log.Fatal("failed to get database connection", err)
	}

	defer func() {
		log.Println("closing database connection")
		_ = conn.Close()
		log.Println("database connection closed")
	}()

	// run the setup required to get the database ready
	// migrations and other setup
	err = dbClient.Setup()
	if err != nil {
		log.Fatal("failed to setup database", err)
	}

	PORT := 3000
	log.Printf("starting server on port %d", PORT)
	api.InitServer(PORT, dckClient, dbClient)
}
