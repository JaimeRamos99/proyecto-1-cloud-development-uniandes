package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/votes"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type VoteHandler struct {
	voteService *votes.Service
}

func NewVoteHandler(db *database.DB) *VoteHandler {
	// Create repository and service
	repo := votes.NewRepository(db)
	service := votes.NewService(repo)

	return &VoteHandler{
		voteService: service,
	}
}

// VoteForVideo handles voting for a video
func (h *VoteHandler) VoteForVideo(c *gin.Context) {
	// Get user ID from JWT claims (guaranteed to exist by AuthMiddleware)
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

	// Call service to vote for video
	err = h.voteService.VoteForVideo(userID, videoID)
	if err != nil {
		errMsg := err.Error()

		// Log the actual error for debugging
		log.Printf("Error voting for video %d by user %d: %v", videoID, userID, err)

		if strings.Contains(errMsg, "video not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Video not found or has been deleted",
			})
		} else if strings.Contains(errMsg, "already voted") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: "You have already voted for this video",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Failed to cast vote. Please try again later.",
			})
		}
		return
	}

	// Get updated vote count
	voteCount, err := h.voteService.GetVideoVoteCount(videoID)
	if err != nil {
		// Log error but don't fail the response since the vote was successful
		voteCount = 0
	}

	// Return success response
	response := &dto.VoteResponse{
		Success:   true,
		Message:   "Vote cast successfully",
		VideoID:   videoID,
		UserID:    userID,
		VotedAt:   time.Now(),
		VoteCount: voteCount,
	}

	c.JSON(http.StatusCreated, response)
}

// UnvoteForVideo handles removing a vote from a video
func (h *VoteHandler) UnvoteForVideo(c *gin.Context) {
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

	// Call service to remove vote
	err = h.voteService.RemoveVote(userID, videoID)
	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "video not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Video not found or has been deleted",
			})
		} else if strings.Contains(errMsg, "vote not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "You have not voted for this video",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: "Failed to remove vote",
			})
		}
		return
	}

	// Get updated vote count
	voteCount, err := h.voteService.GetVideoVoteCount(videoID)
	if err != nil {
		// Log error but don't fail the response since the unvote was successful
		voteCount = 0
	}

	// Return success response
	response := &dto.UnvoteResponse{
		Success:   true,
		Message:   "Vote removed successfully",
		VideoID:   videoID,
		UserID:    userID,
		VoteCount: voteCount,
	}

	c.JSON(http.StatusOK, response)
}
