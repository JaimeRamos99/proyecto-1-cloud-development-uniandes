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

	// Initialize NFS storage provider instead of S3
	nfsProvider, err := providers.NewNFSProvider(&providers.NFSConfig{
		BasePath:   cfg.NFS.BasePath,   // e.g., "/app/shared-files"
		BaseURL:    cfg.NFS.BaseURL,    // e.g., "http://your-web-server.com"
		ServerIP:   cfg.NFS.ServerIP,   // NFS server private IP
		ServerPath: cfg.NFS.ServerPath, // Server-side path
	})
	if err != nil {
		log.Fatalf("Failed to initialize NFS provider: %v", err)
	}

	// Initialize file storage manager
	storageManager := ObjectStorage.NewFileStorageManager(nfsProvider)

	// Initialize job queue instead of SQS
	jobQueue := jobs.NewJobQueue(db, 5*time.Second) // Poll every 5 seconds

	// Initialize video repository
	videoRepo := videos.NewRepository(db)

	// Initialize worker service
	workerService := internal.NewWorkerService(jobQueue, videoRepo, storageManager, &cfg.Retry)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker in a goroutine
	go func() {
		log.Println("Starting worker job processing...")
		
		// Define job handler
		jobHandler := func(job *jobs.VideoJob) error {
			return workerService.ProcessVideoJob(ctx, job)
		}

		// Start polling for jobs
		if err := jobQueue.PollForJobs(ctx, jobHandler); err != nil {
			if err != context.Canceled {
				log.Printf("Worker stopped with error: %v", err)
			}
		}
	}()

	// Start cleanup routine
	go func() {
		cleanupTicker := time.NewTicker(1 * time.Hour)
		defer cleanupTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-cleanupTicker.C:
				// Cleanup jobs older than 24 hours
				if err := jobQueue.CleanupOldJobs(ctx, 24*time.Hour); err != nil {
					log.Printf("Error cleaning up old jobs: %v", err)
				}
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