package videos

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"proyecto1/root/internal/ObjectStorage"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/messaging"
)

type Service struct {
	repo           *Repository
	validator      *FFProbeValidator
	storageManager *ObjectStorage.FileStorageManager
	messageQueue   messaging.MessageQueue
}

func NewService(repo *Repository, storageManager *ObjectStorage.FileStorageManager, messageQueue messaging.MessageQueue) *Service {
	validator := NewFFProbeValidator("/tmp") // Use /tmp for temp files in container
	return &Service{
		repo:           repo,
		validator:      validator,
		storageManager: storageManager,
		messageQueue:   messageQueue,
	}
}

// VideoMetadata represents video file metadata
type VideoMetadata struct {
	Duration   float64 // in seconds
	Width      int
	Height     int
	Size       int64 // file size in bytes
	Format     string
}

// UploadVideo handles the business logic for video upload and validation
func (s *Service) UploadVideo(file *multipart.FileHeader, title string, userID int) (*dto.VideoUploadResponse, error) {
	// Get validation rules
	rules := DefaultValidationRules()
	
	// Perform complete video validation using FFprobe
	_, err := s.validator.ValidateVideo(file, rules)
	if err != nil {
		return nil, fmt.Errorf("video validation failed: %w", err)
	}

	// Create video record in database with metadata
	video := &Video{
		Title:  title,
		Status: StatusUploaded, // Set initial status
		UserID: userID,
	}

	createdVideo, err := s.repo.CreateVideo(video)
	if err != nil {
		return nil, fmt.Errorf("failed to save video record: %w", err)
	}

	// Upload file to S3 using ObjectStorage
	s3Key, err := s.uploadVideoToStorage(file, createdVideo.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video to storage: %w", err)
	} 

	// Send video processing message to message queue
	err = s.sendVideoProcessingMessage(s3Key)
	if err != nil {
		// Log the error but don't fail the upload - the video is already saved
		// In production, you might want to implement retry logic or dead letter queue
		fmt.Printf("Warning: Failed to send message queue message for video %d: %v\n", createdVideo.ID, err)
	}

	// Return success response with S3 information
	response := &dto.VideoUploadResponse{
		ID:         createdVideo.ID,
		Title:      createdVideo.Title,
		Status:     createdVideo.Status,
		UploadedAt: createdVideo.UploadedAt,
		UserID:     createdVideo.UserID,
		S3Key:      s3Key, // Include S3 key in response
	}

	return response, nil
}

// uploadVideoToStorage uploads a video file to S3 and returns the S3 key
func (s *Service) uploadVideoToStorage(file *multipart.FileHeader, videoID, userID int) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read file content into buffer
	fileBuffer, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Generate simple S3 key based on video ID
	s3Key := s.generateS3Key(videoID)

	// Upload to S3 using ObjectStorage
	err = s.storageManager.UploadFile(fileBuffer, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return s3Key, nil
}

// generateS3Key creates an S3 key with "original/" prefix based on video ID
func (s *Service) generateS3Key(videoID int) string {
	// Use "original/" prefix to keep the original file
	return fmt.Sprintf("original/%d.mp4", videoID)
}

// GetVideoDownloadURL generates a presigned URL for video download
func (s *Service) GetVideoDownloadURL(s3Key string) (string, error) {
	url, err := s.storageManager.GetSignedUrl(s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return url, nil
}

// sendVideoProcessingMessage sends a message to message queue for video processing
func (s *Service) sendVideoProcessingMessage(s3Key string) error {
	// Create simple video processing message with just the S3 key
	message := &messaging.VideoProcessingMessage{
		S3Key: s3Key,
	}

	// Send message to message queue with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.messageQueue.SendMessage(ctx, message)
}

// CheckFFProbeInstallation verifies that FFprobe is properly installed
func (s *Service) CheckFFProbeInstallation() error {
	return s.validator.CheckFFProbeInstallation()
}