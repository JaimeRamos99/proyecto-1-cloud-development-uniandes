package handlers

import (
	"net/http"

	"proyecto1/root/internal/ObjectStorage"
	"proyecto1/root/internal/ObjectStorage/providers"
	"proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/messaging"
	messagingProviders "proyecto1/root/internal/messaging/providers"
	"proyecto1/root/internal/videos"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type VideoHandler struct {
	videoService *videos.Service
}

func NewVideoHandler(db *database.DB, cfg *config.Config) *VideoHandler {
	// Create storage manager based on configuration
	storageManager := createStorageManager(cfg)
	
	// Create message queue service
	messageQueue := createMessageQueue(cfg)
	
	// Create repository and service with storage manager and message queue
	repo := videos.NewRepository(db)
	service := videos.NewService(repo, storageManager, messageQueue)
	
	return &VideoHandler{
		videoService: service,
	}
}

// createStorageManager creates S3 storage manager based on configuration
func createStorageManager(cfg *config.Config) *ObjectStorage.FileStorageManager {
	// S3/LocalStack configuration
	s3Config := &providers.S3Config{
		AccessKeyID:     cfg.AWS.AccessKeyID,
		SecretAccessKey: cfg.AWS.SecretAccessKey,
		Region:          cfg.AWS.Region,
		BucketName:      cfg.AWS.S3BucketName,
		EndpointURL:     cfg.AWS.EndpointURL, // LocalStack URL in development, empty for production AWS
	}
	
	// Create S3 provider (works with both LocalStack and real AWS)
	s3Provider, err := providers.NewS3Provider(s3Config)
	if err != nil {
		panic("Failed to create S3 storage provider: " + err.Error())
	}
	
	return ObjectStorage.NewFileStorageManager(s3Provider)
}

// createMessageQueue creates message queue service based on configuration
func createMessageQueue(cfg *config.Config) messaging.MessageQueue {
	// For now, we only have SQS implementation
	// In the future, you can add logic to choose between different providers
	// based on configuration (e.g., cfg.Messaging.Provider)
	messageQueue, err := messagingProviders.NewSQSQueue(&cfg.AWS)
	if err != nil {
		panic("Failed to create message queue service: " + err.Error())
	}
	return messageQueue
}

// UploadVideo handles video upload with authentication and validation
func (h *VideoHandler) UploadVideo(c *gin.Context) {
	// Get user ID from JWT claims (guaranteed to exist by AuthMiddleware)
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// Get video file from form data
	file, err := c.FormFile("video_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Video file is required",
		})
		return
	}

	// Get video title from form data
	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Video title is required",
		})
		return
	}

	// Call service layer for business logic
	response, err := h.videoService.UploadVideo(file, title, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}
