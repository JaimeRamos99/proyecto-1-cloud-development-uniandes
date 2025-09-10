package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"proyecto1/root/internal/ObjectStorage"
	"proyecto1/root/internal/ObjectStorage/providers"
	"proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/messaging"
	messagingProviders "proyecto1/root/internal/messaging/providers"
	"proyecto1/root/internal/videos"
	"proyecto1/root/internal/votes"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type VideoHandler struct {
	videoService *videos.Service
	voteService  *votes.Service
}

func NewVideoHandler(db *database.DB, cfg *config.Config) *VideoHandler {
	// Create storage manager based on configuration
	storageManager := createStorageManager(cfg)
	
	// Create message queue service
	messageQueue := createMessageQueue(cfg)
	
	// Create repository and service with storage manager and message queue
	repo := videos.NewRepository(db)
	service := videos.NewService(repo, storageManager, messageQueue)
	
	// Create vote service
	voteRepo := votes.NewRepository(db)
	voteService := votes.NewService(voteRepo)
	
	return &VideoHandler{
		videoService: service,
		voteService:  voteService,
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

	// Get is_public from form data (required field)
	isPublicStr := c.PostForm("is_public")
	if isPublicStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "is_public field is required",
		})
		return
	}

	// Parse is_public boolean value using strconv.ParseBool (more robust)
	isPublic, parseErr := strconv.ParseBool(strings.ToLower(strings.TrimSpace(isPublicStr)))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "is_public must be a valid boolean value (true/false, 1/0, t/f, T/F, TRUE/FALSE)",
		})
		return
	}

	// Call service layer for business logic
	response, err := h.videoService.UploadVideo(file, title, isPublic, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetVideo retrieves video details with presigned URLs
func (h *VideoHandler) GetVideo(c *gin.Context) {
	// Get user ID from JWT claims
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// Get video ID from URL parameter
	videoIDStr := c.Param("video_id")
	if videoIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Video ID is required",
		})
		return
	}

	// Convert video ID to integer
	videoID, err := strconv.Atoi(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid video ID format",
		})
		return
	}

	// Get video details and URLs from service (with user validation)
	video, originalURL, processedURL, err := h.videoService.GetVideo(videoID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Video not found or not accessible",
		})
		return
	}

	// Get vote count for the video
	voteCount, err := h.voteService.GetVideoVoteCount(videoID)
	if err != nil {
		// Log error but don't fail the response
		log.Printf("Failed to get vote count for video %d: %v", videoID, err)
		voteCount = 0
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
		Votes:        voteCount,
	}

	// Log for debugging (can be removed in production)
	log.Printf("User %d accessed video %d", userID, videoID)

	c.JSON(http.StatusOK, response)
}

// GetUserVideos retrieves all videos for the authenticated user
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	// Get user ID from JWT claims
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// Get all videos for the user from service
	videos, err := h.videoService.GetUserVideos(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve user videos",
		})
		return
	}

	// Update vote counts for all videos
	for _, video := range videos {
		voteCount, err := h.voteService.GetVideoVoteCount(video.VideoID)
		if err != nil {
			// Log error but don't fail the response
			log.Printf("Failed to get vote count for video %d: %v", video.VideoID, err)
			video.Votes = 0
		} else {
			video.Votes = voteCount
		}
	}

	// Log for debugging (can be removed in production)
	log.Printf("User %d retrieved %d videos", userID, len(videos))

	// Return the list of videos
	c.JSON(http.StatusOK, videos)
}

// DeleteVideo handles soft deletion of a video
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	// Get user ID from JWT claims
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// Get video ID from URL parameter
	videoIDStr := c.Param("video_id")
	if videoIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Video ID is required",
		})
		return
	}

	// Convert video ID to integer
	videoID, err := strconv.Atoi(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid video ID format",
		})
		return
	}

	// Call service to delete video
	err = h.videoService.DeleteVideo(videoID, userID)
	if err != nil {
		errMsg := err.Error()
		
		if strings.Contains(errMsg, "video not found or not owned by user") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Video not found or not accessible",
			})
		} else if strings.Contains(errMsg, "public videos cannot be deleted") {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Public videos cannot be deleted. Only private videos can be removed.",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Failed to delete video",
			})
		}
		return
	}

	// Log for debugging (can be removed in production)
	log.Printf("User %d deleted video %d", userID, videoID)

	// Return success response with no content
	c.JSON(http.StatusNoContent, nil)
}

// GetPublicVideos retrieves all public videos without authentication
func (h *VideoHandler) GetPublicVideos(c *gin.Context) {
	// Get all public videos from service (returns []*dto.PublicVideoResponse)
	videos, err := h.videoService.GetPublicVideos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve public videos",
		})
		return
	}

	// ✅ Ensure we return an empty array instead of null
	if videos == nil {
		videos = []*dto.PublicVideoResponse{} // Fixed type
	}

	// Update vote counts for all videos
	for _, video := range videos {
		voteCount, err := h.voteService.GetVideoVoteCount(video.VideoID)
		if err != nil {
			// Log error but don't fail the response
			log.Printf("Failed to get vote count for video %d: %v", video.VideoID, err)
			video.Votes = 0
		} else {
			video.Votes = voteCount
		}
	}

	// Log for debugging (can be removed in production)
	log.Printf("Retrieved %d public videos", len(videos))

	// Return the list of public videos (guaranteed to be an array)
	c.JSON(http.StatusOK, videos)
}
