package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"worker/internal"
	"worker/internal/ObjectStorage"
	"worker/internal/ObjectStorage/providers"
	"worker/internal/config"
	"worker/internal/database"
	messagingProviders "worker/internal/messaging/providers"
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

	// Initialize S3 storage provider
	s3Provider, err := providers.NewS3Provider(&providers.S3Config{
		AccessKeyID:     cfg.AWS.AccessKeyID,
		SecretAccessKey: cfg.AWS.SecretAccessKey,
		Region:          cfg.AWS.Region,
		BucketName:      cfg.AWS.S3BucketName,
		EndpointURL:     cfg.AWS.EndpointURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 provider: %v", err)
	}

	// Initialize file storage manager
	storageManager := ObjectStorage.NewFileStorageManager(s3Provider)

	// Initialize SQS message queue
	messageQueue, err := messagingProviders.NewSQSQueue(&cfg.AWS)
	if err != nil {
		log.Fatalf("Failed to initialize message queue: %v", err)
	}

	// Ensure message queue is closed on exit
	defer func() {
		if err := messageQueue.Close(); err != nil {
			log.Printf("Error closing message queue: %v", err)
		}
	}()

	// Initialize video repository
	videoRepo := videos.NewRepository(db)

	// Initialize worker service
	workerService := internal.NewWorkerService(messageQueue, videoRepo, storageManager)

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
