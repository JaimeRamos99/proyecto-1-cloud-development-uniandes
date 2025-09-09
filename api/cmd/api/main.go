package main

import (
	"log"

	config "proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	httpserver "proyecto1/root/internal/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Ensure database is closed on exit
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}()
	}

	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Create router with configuration and database
	router := httpserver.NewRouter(cfg, db)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting %s v%s on %s (env: %s)", 
		cfg.App.Name, cfg.App.Version, addr, cfg.App.Env)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}