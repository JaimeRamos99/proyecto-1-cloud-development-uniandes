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
	"github.com/golang-jwt/jwt/v5"

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

func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT claims (guaranteed to exist by AuthMiddleware)
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	// Get user profile from service
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "User not found",
		})
		return
	}

	// Create profile response using the same structure as LoginResponse.User
	response := struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		City      string `json:"city"`
		Country   string `json:"country"`
	}{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		City:      user.City,
		Country:   user.Country,
	}

	c.JSON(http.StatusOK, response)
}