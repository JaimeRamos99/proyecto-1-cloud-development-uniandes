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
func (s *Service) UploadVideo(file *multipart.FileHeader, title string, isPublic bool, userID int) (*dto.VideoUploadResponse, error) {
	// Get validation rules
	rules := DefaultValidationRules()
	
	// Perform complete video validation using FFprobe
	_, err := s.validator.ValidateVideo(file, rules)
	if err != nil {
		return nil, fmt.Errorf("video validation failed: %w", err)
	}

	// Create video record in database with metadata
	video := &Video{
		Title:    title,
		Status:   StatusUploaded, // Set initial status
		IsPublic: isPublic,       // Set visibility
		UserID:   userID,
	}

	createdVideo, err := s.repo.CreateVideo(video)
	if err != nil {
		return nil, fmt.Errorf("failed to save video record: %w", err)
	}

	// Upload file to S3 using ObjectStorage
	s3Key, err := s.uploadVideoToStorage(file, createdVideo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video to storage: %w", err)
	} 

	// Send video processing message to message queue
	err = s.sendVideoProcessingMessage(s3Key)
	if err != nil {
		// Log the error but don't fail the upload - the video is already saved
		// In production, you might want to implement retry logic or dead letter queue
		fmt.Printf("Warning: Failed to send message for video %d (S3 key: %s): %v\n", createdVideo.ID, s3Key, err)
	}

	// Return success response with S3 information
	response := &dto.VideoUploadResponse{
		ID:         createdVideo.ID,
		Title:      createdVideo.Title,
		Status:     createdVideo.Status,
		IsPublic:   createdVideo.IsPublic,
		UploadedAt: createdVideo.UploadedAt,
		UserID:     createdVideo.UserID,
		S3Key:      s3Key, // Include S3 key in response
	}

	return response, nil
}

// uploadVideoToStorage uploads a video file to S3 and returns the S3 key
func (s *Service) uploadVideoToStorage(file *multipart.FileHeader, videoID int) (string, error) {
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

// GetVideo retrieves video details and generates presigned URLs (with user validation)
func (s *Service) GetVideo(videoID int, userID int) (*Video, string, string, error) {
	// Get video from database with user ownership validation
	video, err := s.repo.GetVideoByID(videoID, userID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get video: %w", err)
	}

	// Generate presigned URLs for original and processed videos
	originalS3Key := fmt.Sprintf("original/%d.mp4", videoID)
	processedS3Key := fmt.Sprintf("processed/%d.mp4", videoID)

	// Get presigned URL for original video
	originalURL, err := s.storageManager.GetSignedUrl(originalS3Key)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate original video URL: %w", err)
	}

	// Get presigned URL for processed video
	processedURL, err := s.storageManager.GetSignedUrl(processedS3Key)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate processed video URL: %w", err)
	}

	return video, originalURL, processedURL, nil
}

// GetUserVideos retrieves all videos for a user with presigned URLs
func (s *Service) GetUserVideos(userID int) ([]*dto.VideoResponse, error) {
	// Get all videos for the user from database
	videos, err := s.repo.GetVideosByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user videos: %w", err)
	}

	// Convert to response format with presigned URLs
	var responses []*dto.VideoResponse
	for _, video := range videos {
		// Generate presigned URLs for original and processed videos
		originalS3Key := fmt.Sprintf("original/%d.mp4", video.ID)
		processedS3Key := fmt.Sprintf("processed/%d.mp4", video.ID)

		// Get presigned URL for original video
		originalURL, err := s.storageManager.GetSignedUrl(originalS3Key)
		if err != nil {
			// Log error but continue with empty URL
			fmt.Printf("Warning: Failed to generate original video URL for video %d: %v\n", video.ID, err)
			originalURL = ""
		}

		// Get presigned URL for processed video
		processedURL, err := s.storageManager.GetSignedUrl(processedS3Key)
		if err != nil {
			// Log error but continue with empty URL
			fmt.Printf("Warning: Failed to generate processed video URL for video %d: %v\n", video.ID, err)
			processedURL = ""
		}

		// Create response with all required fields
		response := &dto.VideoResponse{
			VideoID:      video.ID,
			Title:        video.Title,
			Status:       video.Status,
			IsPublic:     video.IsPublic,
			UploadedAt:   video.UploadedAt,
			ProcessedAt:  video.ProcessedAt,
			OriginalURL:  originalURL,
			ProcessedURL: processedURL,
			Votes:        0, // Default votes value
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// DeleteVideo performs soft delete on a video (only updates deleted_at, doesn't touch S3)
// Only allows deletion of private videos (is_public = false)
func (s *Service) DeleteVideo(videoID int, userID int) error {
	// Perform soft delete in the database (with public video validation)
	err := s.repo.SoftDeleteVideo(videoID, userID)
	if err != nil {
		// Pass through the specific error messages from repository
		return err
	}

	// Note: We intentionally do NOT delete files from S3 as per requirements
	// The video files remain in storage but the database record is marked as deleted
	
	return nil
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