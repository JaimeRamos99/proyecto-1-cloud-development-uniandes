package handlers

import (
	"log"
	"net/http"
	"strconv"

	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/rankings"

	"github.com/gin-gonic/gin"
)

type RankingHandler struct {
	rankingService *rankings.Service
}

// NewRankingHandler creates a new ranking handler
func NewRankingHandler(db *database.DB) *RankingHandler {
	repo := rankings.NewRepository(db)
	service := rankings.NewService(repo)

	return &RankingHandler{
		rankingService: service,
	}
}

// GetPlayerRankings retrieves player rankings with pagination and filters
func (h *RankingHandler) GetPlayerRankings(c *gin.Context) {
	// Parse pagination parameters
	var pagination rankings.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid pagination parameters: " + err.Error(),
		})
		return
	}

	// Parse filter parameters
	var filters rankings.RankingFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid filter parameters: " + err.Error(),
		})
		return
	}

	// Validate pagination parameters
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 10
	}

	// Get rankings from service
	response, err := h.rankingService.GetPlayerRankings(filters, pagination)
	if err != nil {
		log.Printf("Failed to get player rankings: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve player rankings",
		})
		return
	}

	// Log for debugging (can be removed in production)
	log.Printf("Retrieved %d rankings (page %d, size %d)", len(response.Rankings), pagination.Page, pagination.PageSize)

	c.JSON(http.StatusOK, response)
}

// GetPlayerRanking retrieves a specific player's ranking by user ID
func (h *RankingHandler) GetPlayerRanking(c *gin.Context) {
	// Get user ID from URL parameter
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	// Convert user ID to integer
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid user ID format",
		})
		return
	}

	// Get player ranking from service
	ranking, err := h.rankingService.GetPlayerRanking(userID)
	if err != nil {
		log.Printf("Failed to get player ranking for user %d: %v", userID, err)
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Player ranking not found",
		})
		return
	}

	// Log for debugging (can be removed in production)
	log.Printf("Retrieved ranking for user %d (position %d)", userID, ranking.Ranking)

	c.JSON(http.StatusOK, ranking)
}

// RefreshPlayerRankings manually refreshes the player rankings view
func (h *RankingHandler) RefreshPlayerRankings(c *gin.Context) {
	// Refresh rankings
	err := h.rankingService.RefreshRankings()
	if err != nil {
		log.Printf("Failed to refresh player rankings: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to refresh player rankings",
		})
		return
	}

	log.Println("Player rankings refreshed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Player rankings refreshed successfully",
	})
}
