package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	
	log.Printf("Processing video with S3 key: %s", videoMsg.S3Key)
	
	// Get video record from database using S3 key
	video, err := s.videoRepo.GetVideoByS3Key(videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to get video by S3 key: %w", err)
	}
	
	// Check if video is in the correct status for processing
	if video.Status != videos.StatusUploaded {
		log.Printf("Video %d is not in uploaded status (current: %s), skipping", video.ID, video.Status)
		return nil // Don't return an error - just skip this message
	}
	
	// Download the video file from S3
	log.Printf("Downloading video file: %s", videoMsg.S3Key)
	videoData, err := s.storageManager.DownloadFile(videoMsg.S3Key)
	if err != nil {
		return fmt.Errorf("failed to download video from S3: %w", err)
	}
	
	// Process the video (placeholder for now)
	log.Printf("Processing video file (ID: %d, Size: %d bytes)", video.ID, len(videoData))
	processedData, err := s.processVideo(videoData, video)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}
	
	// Upload the processed video to S3 with "processed/" prefix
	processedS3Key := s.generateProcessedS3Key(videoMsg.S3Key)
	log.Printf("Uploading processed video to S3: %s (original: %s)", processedS3Key, videoMsg.S3Key)
	err = s.storageManager.UploadFile(processedData, processedS3Key)
	if err != nil {
		return fmt.Errorf("failed to upload processed video to S3: %w", err)
	}
	
	// Update video status to processed
	log.Printf("Updating video status to processed: %d", video.ID)
	err = s.videoRepo.UpdateVideoStatus(video.ID, videos.StatusProcessed)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}
	
	log.Printf("Successfully processed video: %d", video.ID)
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

// processVideo performs the actual video processing (placeholder implementation)
func (s *WorkerService) processVideo(videoData []byte, video *videos.Video) ([]byte, error) {
	log.Printf("Processing video ID %d (placeholder implementation)", video.ID)
	
	// TODO: Replace this with actual video processing logic
	// For now, we'll just return the original data unchanged
	// This is where you would add:
	// - Video transcoding
	// - Compression
	// - Watermark addition
	// - Format conversion
	// - Any other video processing operations
	
	// Simulate some processing time
	time.Sleep(2 * time.Second)
	
	log.Printf("Video processing completed for ID %d", video.ID)
	return videoData, nil
}

// Close gracefully shuts down the worker service
func (s *WorkerService) Close() error {
	log.Println("Closing worker service...")
	return s.messageQueue.Close()
}
