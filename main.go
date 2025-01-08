package main

import (
	"code-garden-server/internal/api"
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/docker"
	"database/sql"
	"log"
)

func main() {
	dckClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}
	defer dckClient.Close()

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

	defer func(conn *sql.DB) {
		log.Println("closing database connection")
		err := conn.Close()
		if err != nil {

		}
		log.Println("database connection closed")
	}(conn)

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
