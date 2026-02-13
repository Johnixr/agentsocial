package main

import (
	"fmt"
	"log"

	"agentsocial/internal/api"
	"agentsocial/internal/config"
	"agentsocial/internal/core"
	"agentsocial/internal/db"
)

const version = "0.1.0"

const banner = `
    _                    _   ____             _       _
   / \   __ _  ___ _ __ | |_/ ___|  ___   ___(_) __ _| |
  / _ \ / _' |/ _ \ '_ \| __\___ \ / _ \ / __| |/ _' | |
 / ___ \ (_| |  __/ | | | |_ ___) | (_) | (__| | (_| | |
/_/   \_\__, |\___|_| |_|\__|____/ \___/ \___|_|\__,_|_|
        |___/
`

func main() {
	// Load configuration.
	cfg := config.Load()

	// Print startup banner.
	fmt.Print(banner)
	fmt.Printf("  Version:  %s\n", version)
	fmt.Printf("  Port:     %s\n", cfg.Port)
	fmt.Printf("  Database: %s\n", cfg.SQLitePath)
	fmt.Printf("  Base URL: %s\n", cfg.BaseURL)
	fmt.Println()

	// Initialize database.
	database, err := db.InitDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	log.Println("Database initialized successfully")

	// Create embedding client.
	embClient := core.NewEmbeddingClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIEmbeddingModel,
		cfg.OpenAIEmbeddingDimensions,
	)
	if cfg.OpenAIAPIKey == "" {
		log.Println("WARNING: OpenAI API key is not set. Embedding features will be disabled.")
	} else {
		log.Println("Embedding client initialized")
	}

	// Setup router.
	router := api.SetupRouter(database, cfg, embClient)

	// Start server.
	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
