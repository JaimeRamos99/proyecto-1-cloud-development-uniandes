package dto

// SignupRequest represents the payload for user registration
type SignupRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password1 string `json:"password1" binding:"required,min=6"`
	Password2 string `json:"password2" binding:"required"`
	City      string `json:"city" binding:"required"`
	Country   string `json:"country" binding:"required"`
}

// SignupResponse represents the response for successful user registration
type SignupResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		City      string `json:"city"`
		Country   string `json:"country"`
	} `json:"user"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}
