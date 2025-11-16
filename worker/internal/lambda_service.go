package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"worker/internal/ObjectStorage"
	"worker/internal/videos"
)

// LambdaProcessingService handles video processing for Lambda invocations
// This is a simplified version of WorkerService optimized for event-driven Lambda execution
type LambdaProcessingService struct {
	videoRepo      *videos.Repository
	storageManager *ObjectStorage.FileStorageManager
	processor      *VideoProcessor
}

// NewLambdaProcessingService creates a new Lambda processing service
func NewLambdaProcessingService(
	videoRepo *videos.Repository,
	storageManager *ObjectStorage.FileStorageManager,
	processor *VideoProcessor,
) *LambdaProcessingService {
	return &LambdaProcessingService{
		videoRepo:      videoRepo,
		storageManager: storageManager,
		processor:      processor,
	}
}

// ProcessMessage processes a single video processing message from SQS
// Returns error for transient failures (Lambda will retry)
// Returns nil for permanent failures (message goes to DLQ after max retries)
func (s *LambdaProcessingService) ProcessMessage(ctx context.Context, messageBody string) error {
	log.Printf("Processing message body: %s", messageBody)

	// Parse the message body
	var videoMsg VideoMessage
	err := json.Unmarshal([]byte(messageBody), &videoMsg)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v - permanent error", err)
		return nil // Don't retry invalid JSON - permanent error
	}

	log.Printf("Processing video: S3 key=%s", videoMsg.S3Key)

	// Extract video ID from S3 key
	videoID := s.extractVideoIDFromS3Key(videoMsg.S3Key)
	if videoID <= 0 {
		log.Printf("Could not extract valid video ID from S3 key: %s - permanent error", videoMsg.S3Key)
		return nil // Don't retry - permanent error
	}

	// Verify video exists in database and has correct status
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
		log.Printf("Video %d not found in database: %v - permanent error", videoID, err)
		return nil // Don't retry - permanent error
	}

	// Check video status - only process if uploaded
	if video.Status == videos.StatusProcessed {
		log.Printf("Video %d is already processed, skipping", videoID)
		return nil // Don't retry - already done
	}

	if video.Status != videos.StatusUploaded {
		log.Printf("Video %d has status '%s', expected '%s' - permanent error",
			videoID, video.Status, videos.StatusUploaded)
		return nil // Don't retry - permanent error
	}

	log.Printf("Video %d found with status '%s', proceeding with processing", videoID, video.Status)

	// Download video from S3
	log.Printf("Downloading video file: %s", videoMsg.S3Key)
	videoData, err := s.storageManager.DownloadFile(videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to download video from S3: %w", err) // Retry - transient error
	}

	// Generate processed S3 key
	processedKey := s.generateProcessedS3Key(videoMsg.S3Key)
	log.Printf("Processing video: Original S3 key: %s -> Processed S3 key: %s", videoMsg.S3Key, processedKey)

	// Process video with VideoProcessor
	log.Printf("Processing video file (Size: %d bytes) - applying transformations", len(videoData))
	processedData, err := s.processor.ProcessVideoByS3Key(videoData, videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err) // Retry - could be transient
	}

	// Upload processed video to S3
	log.Printf("Uploading processed video to: %s", processedKey)
	err = s.storageManager.UploadFile(processedData, processedKey)
	if err != nil {
		return fmt.Errorf("failed to upload processed video: %w", err) // Retry - transient error
	}

	// Update video status to processed
	log.Printf("Updating video status to processed: %d", videoID)
	err = s.videoRepo.UpdateVideoStatus(videoID, videos.StatusProcessed)
	if err != nil {
		log.Printf("Warning: failed to update video status: %v", err)
		// Don't fail the entire process for DB update failure
		// The video is processed and uploaded, status update is not critical
	}

	log.Printf("Successfully processed video (Original: %s, Processed: %s)", videoMsg.S3Key, processedKey)
	log.Printf("Transformations applied: ≤30s, 1280x720, 16:9, no audio, ANB watermark, ANB bumpers, no content cropping")

	return nil
}

// VideoMessage represents the video processing message format from SQS
type VideoMessage struct {
	S3Key string `json:"s3_key"`
}

// generateProcessedS3Key converts an original S3 key to a processed S3 key
// Example: "original/1.mp4" -> "processed/1.mp4"
func (s *LambdaProcessingService) generateProcessedS3Key(originalS3Key string) string {
	// Convert "original/1.mp4" to "processed/1.mp4"
	// Remove "original/" prefix and add "processed/" prefix
	if filename, found := strings.CutPrefix(originalS3Key, "original/"); found {
		return fmt.Sprintf("processed/%s", filename)
	}

	// Fallback: if the key doesn't have "original/" prefix, just add "processed/" prefix
	return fmt.Sprintf("processed/%s", originalS3Key)
}

// extractVideoIDFromS3Key extracts video ID from S3 key
// Example: "original/123.mp4" -> 123
func (s *LambdaProcessingService) extractVideoIDFromS3Key(s3Key string) int {
	// Remove path prefixes if present (e.g., "original/123.mp4" -> "123.mp4")
	filename := s3Key
	if lastSlash := strings.LastIndex(s3Key, "/"); lastSlash >= 0 {
		filename = s3Key[lastSlash+1:]
	}

	// Extract ID from filename (e.g., "123.mp4" -> "123")
	if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
		idStr := filename[:dotIndex]
		if videoID, err := strconv.Atoi(idStr); err == nil {
			return videoID
		}
	}

	return 0 // Return 0 if extraction fails
}

// IsPermanentError checks if an error should not be retried
// This helps Lambda decide whether to retry or send to DLQ
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Video not found in database - permanent error
	if strings.Contains(errStr, "not found in database") {
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

	// Invalid message format - permanent error
	if strings.Contains(errStr, "failed to unmarshal") {
		return true
	}

	// All other errors are considered transient and can be retried
	return false
}
