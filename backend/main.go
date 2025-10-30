package main

import (
	"dbpedia-server/client"
	"dbpedia-server/config"
	"dbpedia-server/handlers"
	"dbpedia-server/server"
	"log"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize DBpedia client
	dbpediaClient := client.NewDBpediaClient(cfg.DBpediaEndpoint)

	// Initialize handlers
	handler := handlers.NewHandler(dbpediaClient)

	// Initialize server
	srv := server.NewServer(handler, server.Config{
		Port:        cfg.ServerPort,
		ReleaseMode: cfg.ReleaseMode,
	})

	// Setup routes
	srv.SetupRoutes()

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
