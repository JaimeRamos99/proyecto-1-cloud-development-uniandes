package handlers

import (
	"net/http"
	"time"

	"proyecto1/root/internal/config"
	"proyecto1/root/internal/database"
	"proyecto1/root/internal/http/dto"
	"proyecto1/root/internal/http/session"
	"proyecto1/root/internal/users"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *users.Service
	sessions    *session.InMemorySessionStore
}

// NewAuthHandler creates an AuthHandler with a shared session store
func NewAuthHandler(db *database.DB, cfg *config.Config, sessionStore *session.InMemorySessionStore) *AuthHandler {
	repo := users.NewRepository(db)
	service := users.NewService(repo, cfg)
	return &AuthHandler{
		userService: service,
		sessions:    sessionStore,
	}
}

// Signup handles user registration HTTP request
func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest

	// Bind JSON to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
		return
	}

	// Call service layer for business logic
	response, err := h.userService.Signup(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// Bind JSON to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
		})
		return
	}

	// Call service layer for business logic
	response, err := h.userService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout invalidates the provided JWT by adding it to a revoked set (simulated server-side invalidation).
func (h *AuthHandler) Logout(c *gin.Context) {
	// Typical logout for stateless JWT is performed on client by discarding the token.
	// When server-side invalidation is required, a token blacklist/denylist is necessary.
	// Here we simulate that using an in-memory map.
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization header required"})
		return
	}

	// Expecting format: "Bearer <token>"
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header"})
		return
	}

	token := authHeader[len(prefix):]
	h.sessions.RevokeToken(token, time.Now().Add(24*time.Hour))
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// GetProfile handles getting user profile information
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "unauthorized",
		})
		return
	}

	// Convert userID to int
	id, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "invalid user ID format",
		})
		return
	}

	// Get user profile from service
	profile, err := h.userService.GetProfile(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}