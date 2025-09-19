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

// Profile returns the currently authenticated user's profile
func (h *AuthHandler) Profile(c *gin.Context) {
    userIDVal, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user not authenticated"})
        return
    }
    userID, ok := userIDVal.(int)
    if !ok {
        c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid user id in context"})
        return
    }
    // Fetch user from DB
    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to fetch user"})
        return
	}
    // Return user profile 
    response := dto.SignupResponse{
        ID:        user.ID,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Email:     user.Email,
        City:      user.City,
        Country:   user.Country,
    }

    c.JSON(http.StatusOK, response)
}

