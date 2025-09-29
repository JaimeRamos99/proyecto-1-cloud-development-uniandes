package internal

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"worker/internal/ObjectStorage"
	"worker/internal/config"
	"worker/internal/jobs"
	"worker/internal/videos"
)

// WorkerService handles video processing tasks from the database job queue
type WorkerService struct {
	jobQueue       *jobs.JobQueue
	videoRepo      *videos.Repository
	storageManager *ObjectStorage.FileStorageManager
	retryConfig    *config.RetryConfig
}

// NewWorkerService creates a new worker service
func NewWorkerService(
	jobQueue *jobs.JobQueue,
	videoRepo *videos.Repository,
	storageManager *ObjectStorage.FileStorageManager,
	retryConfig *config.RetryConfig,
) *WorkerService {
	return &WorkerService{
		jobQueue:       jobQueue,
		videoRepo:      videoRepo,
		storageManager: storageManager,
		retryConfig:    retryConfig,
	}
}

// ProcessMessages starts the worker to poll for jobs from the database
func (s *WorkerService) ProcessMessages(ctx context.Context) error {
	log.Println("Worker started, polling for jobs from database...")

	// Define the job handler function
	jobHandler := func(job *jobs.VideoJob) error {
		return s.processJobWithRetry(ctx, job)
	}

	// Start polling for jobs
	return s.jobQueue.PollForJobs(ctx, jobHandler)
}

// processJobWithRetry processes a job with exponential backoff retry logic
func (s *WorkerService) processJobWithRetry(ctx context.Context, job *jobs.VideoJob) error {
	if !s.retryConfig.EnableBackoff {
		// If retry is disabled, just process once
		return s.processJob(ctx, job)
	}

	var lastErr error
	maxRetries := s.retryConfig.MaxRetries
	baseDelay := time.Duration(s.retryConfig.BaseDelay) * time.Second
	maxDelay := time.Duration(s.retryConfig.MaxDelay) * time.Second

	// Use job's attempt count as the starting point
	currentAttempt := job.Attempts

	for attempt := currentAttempt; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(attempt-2))) * baseDelay
			if delay > maxDelay {
				delay = maxDelay
			}

			log.Printf("Job %d failed on attempt %d, retrying after %v", job.ID, attempt, delay)

			// Sleep with context cancellation support
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to retry
			}
		}

		// Attempt to process the job
		lastErr = s.processJob(ctx, job)
		if lastErr == nil {
			if attempt > 1 {
				log.Printf("Job %d succeeded on retry attempt %d", job.ID, attempt)
			}
			return nil // Success
		}

		// Check if this is a permanent error that shouldn't be retried
		if s.isPermanentError(lastErr) {
			log.Printf("Job %d failed with permanent error, not retrying: %v", job.ID, lastErr)
			return lastErr
		}

		log.Printf("Job %d attempt %d failed (will retry): %v", job.ID, attempt, lastErr)
	}

	log.Printf("Job %d failed after %d attempts, giving up: %v", job.ID, maxRetries, lastErr)
	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// isPermanentError determines if an error is permanent and shouldn't be retried
func (s *WorkerService) isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Video not found in database - permanent error
	if strings.Contains(errStr, "not found in database") || strings.Contains(errStr, "not found") {
		return true
	}

	// Video already processed - permanent error
	if strings.Contains(errStr, "already processed") {
		return true
	}

	// Invalid video format - permanent error
	if strings.Contains(errStr, "invalid video format") || strings.Contains(errStr, "unsupported format") {
		return true
	}

	// File not found - permanent error
	if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "file not found") {
		return true
	}

	// All other errors are considered temporary and can be retried
	return false
}

// processJob processes a single video processing job
func (s *WorkerService) processJob(ctx context.Context, job *jobs.VideoJob) error {
	log.Printf("Processing job %d for video %d", job.ID, job.VideoID)

	// Get video from database
	video, err := s.videoRepo.GetVideoByID(job.VideoID)
	if err != nil {
		errMsg := fmt.Sprintf("video not found in database: %v", err)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Check video status - only process if uploaded
	if video.Status == videos.StatusProcessed {
		log.Printf("Video %d is already processed, skipping job %d", job.VideoID, job.ID)
		// Mark job as completed since video is already processed
		s.jobQueue.CompleteJob(ctx, job.ID, job.FilePath)
		return nil
	}

	if video.Status != videos.StatusUploaded {
		log.Printf("Video %d has status '%s', expected '%s' for processing - skipping job %d", 
			job.VideoID, video.Status, videos.StatusUploaded, job.ID)
		// Mark as failed since video is in wrong state
		errMsg := fmt.Sprintf("video status is '%s', expected '%s'", video.Status, videos.StatusUploaded)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("Video %d found with status '%s', proceeding with processing", job.VideoID, video.Status)

	// Get the file path from NFS storage
	inputFilePath := s.storageManager.GetFilePath(filepath.Base(job.FilePath), "uploads")
	
	// Check if file exists
	if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("source file not found: %s", inputFilePath)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("Processing video file: %s", inputFilePath)

	// Read the video file
	videoData, err := os.ReadFile(inputFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read video file: %v", err)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Generate output filename
	outputFileName := s.generateProcessedFileName(video.Filename)
	outputFilePath := s.storageManager.GetFilePath(outputFileName, "processed")

	log.Printf("Processing video file (Size: %d bytes) - applying transformations", len(videoData))
	
	// Process video with VideoProcessor
	processor := NewVideoProcessor()
	processedData, err := processor.ProcessVideoByFilename(videoData, video.Filename)
	if err != nil {
		errMsg := fmt.Sprintf("failed to process video: %v", err)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFilePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create output directory: %v", err)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Write processed video to NFS
	log.Printf("Writing processed video to: %s", outputFilePath)
	if err := os.WriteFile(outputFilePath, processedData, 0644); err != nil {
		errMsg := fmt.Sprintf("failed to write processed video: %v", err)
		s.jobQueue.FailJob(ctx, job.ID, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Update job status to completed
	err = s.jobQueue.CompleteJob(ctx, job.ID, outputFileName)
	if err != nil {
		log.Printf("Warning: failed to mark job as completed: %v", err)
	}

	// Update video status to processed
	log.Printf("Updating video status to processed: %d", job.VideoID)
	err = s.videoRepo.UpdateVideoStatus(job.VideoID, videos.StatusProcessed)
	if err != nil {
		log.Printf("Warning: failed to update video status: %v", err)
		// Don't fail the entire process for DB update failure
	}

	log.Printf("Successfully processed job %d (Input: %s, Output: %s)",
		job.ID, inputFilePath, outputFilePath)
	log.Printf("Transformations applied: ≤30s, 1280x720, 16:9, no audio, ANB watermark, ANB bumpers, no content cropping")
	
	return nil
}

// generateProcessedFileName converts an original filename to a processed filename
func (s *WorkerService) generateProcessedFileName(originalFilename string) string {
	// Get file extension
	ext := filepath.Ext(originalFilename)
	
	// Get filename without extension
	nameWithoutExt := strings.TrimSuffix(originalFilename, ext)
	
	// Add timestamp and processed suffix
	timestamp := time.Now().Format("20060102_150405")
	
	return fmt.Sprintf("%s_processed_%s%s", nameWithoutExt, timestamp, ext)
}

// Close gracefully shuts down the worker service
func (s *WorkerService) Close() error {
	log.Println("Closing worker service...")
	return nil
}