package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"worker/internal"
	"worker/internal/ObjectStorage"
	"worker/internal/ObjectStorage/providers"
	"worker/internal/config"
	"worker/internal/database"
	"worker/internal/videos"
)

// Global variables for connection reuse across Lambda invocations
// These are initialized once during cold start and reused for subsequent invocations
var (
	processingService *internal.LambdaProcessingService
	cfg               *config.Config
)

// init runs once when Lambda container starts (cold start)
// Initializes all database connections and AWS clients for reuse
func init() {
	log.Println("Lambda cold start - initializing connections...")

	// Load configuration from environment variables
	cfg = config.Load()
	log.Printf("Starting %s v%s (env: %s)", cfg.App.Name, cfg.App.Version, cfg.App.Env)

	// Initialize database connection (reused across invocations)
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connection initialized")

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
	log.Println("S3 storage provider initialized")

	// Initialize components
	storageManager := ObjectStorage.NewFileStorageManager(s3Provider)
	videoRepo := videos.NewRepository(db)
	processor := internal.NewVideoProcessor()

	// Initialize processing service (reused across invocations)
	processingService = internal.NewLambdaProcessingService(videoRepo, storageManager, processor)

	log.Println("Lambda initialization complete - ready to process messages")
}

// handler processes SQS events from Lambda
// This function is called by AWS Lambda runtime for each invocation
func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("Lambda invoked with %d messages", len(sqsEvent.Records))

	// Process each record in the batch
	for _, record := range sqsEvent.Records {
		if err := processRecord(ctx, record); err != nil {
			log.Printf("Error processing message %s: %v", record.MessageId, err)
			// Return error to trigger Lambda retry
			// Lambda will automatically retry failed messages
			return err
		}
	}

	log.Printf("Successfully processed %d messages", len(sqsEvent.Records))
	return nil
}

// processRecord processes a single SQS record using the processing service
func processRecord(ctx context.Context, record events.SQSMessage) error {
	log.Printf("Processing message ID: %s", record.MessageId)

	// Delegate to processing service
	// Service returns nil for permanent errors (no retry)
	// Service returns error for transient errors (Lambda retries)
	return processingService.ProcessMessage(ctx, record.Body)
}

// main is the entry point for the Lambda function
func main() {
	// Start the Lambda handler
	lambda.Start(handler)
}
