package users

import (
	"errors"
	"strconv"

	"proyecto1/root/internal/auth"
	"proyecto1/root/internal/config"
	"proyecto1/root/internal/http/dto"
)

type Service struct {
	repo         *Repository
	tokenManager *auth.TokenManager
	jwtConfig    *config.JWTConfig
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	tokenManager := &auth.TokenManager{
		Secret: []byte(cfg.JWT.Secret),
		Issuer: cfg.JWT.Issuer,
	}
	return &Service{
		repo:         repo,
		tokenManager: tokenManager,
		jwtConfig:    &cfg.JWT,
	}
}

// Signup handles the business logic for user registration
func (s *Service) Signup(req dto.SignupRequest) (*dto.SignupResponse, error) {
	// Validate passwords match
	if req.Password1 != req.Password2 {
		return nil, errors.New("passwords do not match")
	}

	// Check if email already exists
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		return nil, errors.New("failed to validate email")
	}
	
	if exists {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password1)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create user model
	user := &User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		City:         req.City,
		Country:      req.Country,
	}

	// Save user to database
	createdUser, err := s.repo.CreateUser(user)
	if err != nil {
		if err.Error() == "email already exists" {
			return nil, errors.New("email already exists")
		}
		return nil, errors.New("failed to create user")
	}

	// Return success response
	response := &dto.SignupResponse{
		ID:        createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
		City:      createdUser.City,
		Country:   createdUser.Country,
	}

	return response, nil
}

// Login handles the business logic for user login
func (s *Service) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	customClaims := map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	}
	token, err := s.tokenManager.CreateToken(strconv.Itoa(user.ID), s.jwtConfig.Expiration, customClaims)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Create response
	response := &dto.LoginResponse{
		Token: token,
	}
	response.User.ID = user.ID
	response.User.FirstName = user.FirstName
	response.User.LastName = user.LastName
	response.User.Email = user.Email
	response.User.City = user.City
	response.User.Country = user.Country

	return response, nil
}
