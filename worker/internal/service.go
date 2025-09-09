package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"worker/internal/ObjectStorage"
	"worker/internal/messaging"
	"worker/internal/videos"
)

// WorkerService handles video processing tasks from the message queue
type WorkerService struct {
	messageQueue   messaging.MessageQueue
	videoRepo      *videos.Repository
	storageManager *ObjectStorage.FileStorageManager
}

// NewWorkerService creates a new worker service
func NewWorkerService(
	messageQueue messaging.MessageQueue,
	videoRepo *videos.Repository,
	storageManager *ObjectStorage.FileStorageManager,
) *WorkerService {
	return &WorkerService{
		messageQueue:   messageQueue,
		videoRepo:      videoRepo,
		storageManager: storageManager,
	}
}

// ProcessMessages starts the worker to process messages from the queue
func (s *WorkerService) ProcessMessages(ctx context.Context) error {
	log.Println("Worker started, listening for messages...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping due to context cancellation")
			return ctx.Err()
		default:
			err := s.processMessageBatch(ctx)
			if err != nil {
				log.Printf("Error processing message batch: %v", err)
				// Continue processing even if there's an error
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// processMessageBatch fetches and processes a batch of messages
func (s *WorkerService) processMessageBatch(ctx context.Context) error {
	// Use long polling with maximum wait time (20 seconds for SQS)
	messages, err := s.messageQueue.ReceiveMessages(ctx, 10, 20)
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}
	
	if len(messages) == 0 {
		// No messages received, this is normal with long polling
		return nil
	}
	
	log.Printf("Received %d messages for processing", len(messages))
	
	// Process each message
	for _, msg := range messages {
		err := s.processMessage(ctx, msg)
		if err != nil {
			log.Printf("Error processing message %s: %v", msg.MessageID, err)
			// Continue processing other messages even if one fails
			continue
		}
		
		// Delete the message from the queue after successful processing
		err = s.messageQueue.DeleteMessage(ctx, msg.ReceiptHandle)
		if err != nil {
			log.Printf("Error deleting message %s: %v", msg.MessageID, err)
			// Continue processing - the message will be redelivered later
		}
	}
	
	return nil
}

// processMessage processes a single video processing message
func (s *WorkerService) processMessage(ctx context.Context, msg *messaging.ReceivedMessage) error {
	log.Printf("Processing message: %s", msg.MessageID)
	
	// Parse the message body
	var videoMsg messaging.VideoProcessingMessage
	err := json.Unmarshal([]byte(msg.Body), &videoMsg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Extract video ID and verify status before processing
	videoID := s.extractVideoIDFromS3Key(videoMsg.S3Key)
	if videoID <= 0 {
		log.Printf("Could not extract valid video ID from S3 key: %s - skipping processing", videoMsg.S3Key)
		return nil // Skip processing, message will be deleted
	}

	// Get video from database - MUST exist for processing
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
		log.Printf("Video %d not found in database: %v - skipping processing", videoID, err)
		return nil // Skip processing if video doesn't exist in database
	}

	// Check video status - only process if uploaded
	if video.Status == videos.StatusProcessed {
		log.Printf("Video %d is already processed, skipping processing", videoID)
		return nil // Skip processing if already processed
	}
	
	if video.Status != videos.StatusUploaded {
		log.Printf("Video %d has status '%s', expected '%s' for processing - skipping", videoID, video.Status, videos.StatusUploaded)
		return nil // Skip processing if not in uploaded status
	}
	
	log.Printf("Video %d found with status '%s', proceeding with processing", videoID, video.Status)

	log.Printf("Downloading video file: %s", videoMsg.S3Key)
	videoData, err := s.storageManager.DownloadFile(videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to download video from S3: %w", err)
	}
	
	// Generate processed key (API sends "original/123.mp4", we want "processed/123.mp4")
	processedKey := s.generateProcessedS3Key(videoMsg.S3Key)
	
	log.Printf("Processing video: Original S3 key: %s -> Processed S3 key: %s", videoMsg.S3Key, processedKey)
	
	// Process video with VideoProcessor
	log.Printf("Processing video file (Size: %d bytes) - applying transformations", len(videoData))
	processor := NewVideoProcessor()
	processedData, err := processor.ProcessVideoByS3Key(videoData, videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}
	
	// Upload processed video to processed/ location (keeping original in original/)
	log.Printf("Uploading processed video to: %s", processedKey)
	err = s.storageManager.UploadFile(processedData, processedKey)
	if err != nil {
		return fmt.Errorf("failed to upload processed video: %w", err)
	}
	
	// Update video status to processed - use already extracted videoID
	if videoID > 0 {
		log.Printf("Updating video status to processed: %d", videoID)
		err = s.videoRepo.UpdateVideoStatus(videoID, videos.StatusProcessed)
		if err != nil {
			log.Printf("Warning: failed to update video status: %v", err)
			// Don't fail the entire process for DB update failure
		}
	}
	
	log.Printf("Successfully processed video (Original: %s, Processed: %s)", 
		videoMsg.S3Key, processedKey)
	log.Printf("Transformations applied: ≤30s, 1280x720, 16:9, no audio, ANB watermark, ANB bumpers, no content cropping")
	return nil
}

// generateProcessedS3Key converts an original S3 key to a processed S3 key
func (s *WorkerService) generateProcessedS3Key(originalS3Key string) string {
	// Convert "original/1.mp4" to "processed/1.mp4"
	// Remove "original/" prefix and add "processed/" prefix
	if filename, found := strings.CutPrefix(originalS3Key, "original/"); found {
		return fmt.Sprintf("processed/%s", filename)
	}
	
	// Fallback: if the key doesn't have "original/" prefix, just add "processed/" prefix
	return fmt.Sprintf("processed/%s", originalS3Key)
}

// extractVideoIDFromS3KeyForDB extracts video ID from S3 key only for database updates
func (s *WorkerService) extractVideoIDFromS3Key(s3Key string) int {
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
	
	return 0 // Return 0 if extraction fails - DB update will be skipped
}
  
// Close gracefully shuts down the worker service
func (s *WorkerService) Close() error {
	log.Println("Closing worker service...")
	return s.messageQueue.Close()
}
