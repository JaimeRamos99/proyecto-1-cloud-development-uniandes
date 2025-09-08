package handlers

import (
	"net/http"

	"proyecto1/root/internal/ObjectStorage"
	"proyecto1/root/internal/ObjectStorage/providers"
	"proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
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
	
	// Create repository and service with storage manager
	repo := videos.NewRepository(db)
	service := videos.NewService(repo, storageManager)
	
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
