package videos

import (
	"fmt"
	"io"
	"mime/multipart"

	"proyecto1/root/internal/ObjectStorage"
	"proyecto1/root/internal/http/dto"
)

type Service struct {
	repo           *Repository
	validator      *FFProbeValidator
	storageManager *ObjectStorage.FileStorageManager
}

func NewService(repo *Repository, storageManager *ObjectStorage.FileStorageManager) *Service {
	validator := NewFFProbeValidator("/tmp") // Use /tmp for temp files in container
	return &Service{
		repo:           repo,
		validator:      validator,
		storageManager: storageManager,
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

// generateS3Key creates a simple S3 key based on video ID
func (s *Service) generateS3Key(videoID int) string {
	// Simple approach: {id}.mp4
	return fmt.Sprintf("%d.mp4", videoID)
}

// GetVideoDownloadURL generates a presigned URL for video download
func (s *Service) GetVideoDownloadURL(s3Key string) (string, error) {
	url, err := s.storageManager.GetSignedUrl(s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return url, nil
}

// CheckFFProbeInstallation verifies that FFprobe is properly installed
func (s *Service) CheckFFProbeInstallation() error {
	return s.validator.CheckFFProbeInstallation()
}