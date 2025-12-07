package dto

// Request DTOs

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

// Response DTOs

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// ProtectedResponse represents a protected resource response
type ProtectedResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

// UserListResponse represents a user in list responses (without password)
type UserListResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}
