package handlers

import (
	"net/http"

	"proyecto1/root/internal/database"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db           *database.DB
	videoHandler *VideoHandler
}

// NewHealthHandler creates a health handler for system status checks
func NewHealthHandler(db *database.DB, videoHandler *VideoHandler) *HealthHandler {
	return &HealthHandler{
		db:           db,
		videoHandler: videoHandler,
	}
}

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Message  string            `json:"message,omitempty"`
}

// Health performs comprehensive health check of all services
func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]string)
	overallStatus := "healthy"
	var messages []string

	// Check database connection
	if err := h.db.Ping(); err != nil {
		services["database"] = "unhealthy"
		overallStatus = "unhealthy"
		messages = append(messages, "Database connection failed")
	} else {
		services["database"] = "healthy"
	}

	// Check FFprobe installation
	if h.videoHandler != nil && h.videoHandler.videoService != nil {
		if err := h.videoHandler.videoService.CheckFFProbeInstallation(); err != nil {
			services["ffprobe"] = "unhealthy"
			overallStatus = "unhealthy"
			messages = append(messages, "FFprobe not available for video validation")
		} else {
			services["ffprobe"] = "healthy"
		}
	} else {
		services["ffprobe"] = "unknown"
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:   overallStatus,
		Services: services,
	}

	if len(messages) > 0 {
		response.Message = messages[0] // Return first error message
	}

	c.JSON(statusCode, response)
}
