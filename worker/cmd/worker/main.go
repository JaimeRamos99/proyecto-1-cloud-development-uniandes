package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"worker/internal"
	"worker/internal/ObjectStorage"
	"worker/internal/ObjectStorage/providers"
	"worker/internal/config"
	"worker/internal/database"
	"worker/internal/jobs"
	"worker/internal/videos"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	log.Printf("Starting %s v%s (env: %s)", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	// Initialize database connection
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Ensure database is closed on exit
	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}
	}()

	// Initialize NFS storage provider (CHANGED FROM S3)
	nfsProvider, err := providers.NewNFSProvider(&providers.NFSConfig{
		BasePath: "/app/shared-files",
		BaseURL:  cfg.NFS.BaseURL, // Add this to your config
	})
	if err != nil {
		log.Fatalf("Failed to initialize NFS provider: %v", err)
	}

	// Initialize file storage manager
	storageManager := ObjectStorage.NewFileStorageManager(nfsProvider)

	// Initialize job queue (CHANGED FROM SQS)
	jobQueue := jobs.NewJobQueue(db, 5*time.Second)

	// Initialize video repository
	videoRepo := videos.NewRepository(db)

	// Initialize worker service
	workerService := internal.NewWorkerService(jobQueue, videoRepo, storageManager, &cfg.Retry)

	// Ensure worker service is closed on exit
	defer func() {
		if err := workerService.Close(); err != nil {
			log.Printf("Error closing worker service: %v", err)
		}
	}()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker in a goroutine
	go func() {
		log.Println("Starting worker message processing...")
		if err := workerService.ProcessMessages(ctx); err != nil {
			if err != context.Canceled {
				log.Printf("Worker stopped with error: %v", err)
			}
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down worker...", sig)

	// Cancel context to stop worker
	cancel()

	log.Println("Worker shutdown complete")
}