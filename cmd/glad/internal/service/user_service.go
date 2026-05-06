package service

import (
	"time"

	"github.com/hackmajoris/glad-stack/cmd/glad/internal/database"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/dto"
	apperrors "github.com/hackmajoris/glad-stack/cmd/glad/internal/errors"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/pkg/auth"
	pkgerrors "github.com/hackmajoris/glad-stack/pkg/errors"
	"github.com/hackmajoris/glad-stack/pkg/logger"
)

// Re-export domain errors for convenience in handler layer
var (
	ErrUserExists         = apperrors.ErrUserExists
	ErrUserNotFound       = apperrors.ErrUserNotFound
	ErrInvalidCredentials = apperrors.ErrInvalidCredentials
	ErrInvalidUsername    = apperrors.ErrInvalidUsername
	ErrInvalidName        = apperrors.ErrInvalidName
	ErrInvalidPassword    = apperrors.ErrInvalidPassword
)

// UserService handles user business logic
type UserService struct {
	repo         database.UserRepository
	tokenService *auth.TokenService
}

// NewUserService creates a new UserService
func NewUserService(repo database.UserRepository, tokenService *auth.TokenService) *UserService {
	return &UserService{
		repo:         repo,
		tokenService: tokenService,
	}
}

// RegisterResult contains the result of a registration
type RegisterResult struct {
	Username string
}

// Register registers a new user
func (s *UserService) Register(username, name, password string) (*RegisterResult, error) {
	log := logger.WithComponent("service").With("operation", "Register", "username", username)
	start := time.Now()

	log.Info("Processing registration request")

	// Check if user already exists
	exists, err := s.repo.UserExists(username)
	if err != nil {
		log.Error("Failed to check user existence", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}
	if exists {
		log.Info("Registration attempt with existing username", "duration", time.Since(start))
		return nil, ErrUserExists
	}

	// Create new user
	user, err := models.NewUser(username, name, password)
	if err != nil {
		log.Error("Failed to create user model", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Save user to database
	if err := s.repo.CreateUser(user); err != nil {
		log.Error("Failed to save user to database", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("User registered successfully", "duration", time.Since(start))
	return &RegisterResult{Username: username}, nil
}

// CreateUserFromCognito creates a DynamoDB user profile from Cognito user data
// DEPRECATED: This method is no longer used. User profile creation is now handled by the
// Cognito Post Confirmation Lambda trigger (cmd/cognito-triggers/post-confirmation).
// The trigger automatically creates the DynamoDB profile when a user confirms their email.
// This method is kept for backward compatibility and potential migration scenarios.
func (s *UserService) CreateUserFromCognito(username, email string) (*models.User, error) {
	log := logger.WithComponent("service").With("operation", "CreateUserFromCognito", "username", username)
	start := time.Now()

	log.Info("Creating user profile from Cognito data")

	// Check if user already exists (shouldn't happen, but safe to check)
	exists, err := s.repo.UserExists(username)
	if err != nil {
		log.Error("Failed to check user existence", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}
	if exists {
		log.Info("User profile already exists, skipping creation", "duration", time.Since(start))
		// Return existing user instead of error
		return s.repo.GetUser(username)
	}

	// Create minimal user profile from Cognito data
	// Note: No password hash since Cognito handles authentication
	now := time.Now()
	user := &models.User{
		Username:     username,
		Name:         username, // Default to username, user can update later
		Email:        email,
		PasswordHash: "", // Empty - Cognito handles passwords
		CreatedAt:    now,
		UpdatedAt:    now,
		EntityType:   "User",
	}

	// Set DynamoDB keys
	user.SetKeys()

	// Save user to database
	if err := s.repo.CreateUser(user); err != nil {
		log.Error("Failed to create user profile in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("User profile created successfully from Cognito data", "duration", time.Since(start))
	return user, nil
}

// LoginResult contains the result of a login
type LoginResult struct {
	AccessToken string
	TokenType   string
}

// Login authenticates a user and returns a token
func (s *UserService) Login(username, password string) (*LoginResult, error) {
	log := logger.WithComponent("service").With("operation", "Login", "username", username)
	start := time.Now()

	log.Info("Processing login request")

	// Get user from database
	user, err := s.repo.GetUser(username)
	if err != nil {
		if pkgerrors.Is(err, apperrors.ErrUserNotFound) {
			log.Info("Login attempt with non-existent username", "duration", time.Since(start))
			return nil, apperrors.ErrInvalidCredentials
		}
		log.Error("Failed to retrieve user for login", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Validate password
	if !user.ValidatePassword(password) {
		log.Info("Login attempt with invalid password", "duration", time.Since(start))
		return nil, apperrors.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.tokenService.GenerateToken(user)
	if err != nil {
		log.Error("Failed to generate JWT token", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("User logged in successfully", "duration", time.Since(start))
	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}

// UpdateUser updates a user's profile
func (s *UserService) UpdateUser(username string, name *string, password *string) error {
	log := logger.WithComponent("service").With("operation", "UpdateUser", "username", username)
	start := time.Now()

	log.Info("Processing update request")

	// Get current user
	user, err := s.repo.GetUser(username)
	if err != nil {
		log.Error("Failed to get user", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	// Update user fields
	if name != nil {
		if err := user.UpdateName(*name); err != nil {
			log.Error("Failed to update user name", "error", err.Error(), "duration", time.Since(start))
			return err
		}
	}

	if password != nil {
		if err := user.UpdatePassword(*password); err != nil {
			log.Error("Failed to update user password", "error", err.Error(), "duration", time.Since(start))
			return err
		}
	}

	// Save updated user
	if err := s.repo.UpdateUser(user); err != nil {
		log.Error("Failed to save user", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("User updated successfully", "duration", time.Since(start))
	return nil
}

// GetUser retrieves a user by username
func (s *UserService) GetUser(username string) (*models.User, error) {
	return s.repo.GetUser(username)
}

// ListUsers retrieves all users
func (s *UserService) ListUsers() ([]dto.UserListResponse, error) {
	log := logger.WithComponent("service").With("operation", "ListUsers")
	start := time.Now()

	log.Info("Processing list users request")

	users, err := s.repo.ListUsers()
	if err != nil {
		log.Error("Failed to retrieve users", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Convert to list items (without sensitive data)
	result := make([]dto.UserListResponse, len(users))
	for i, user := range users {
		result[i] = dto.UserListResponse{
			Username: user.Username,
			Name:     user.Name,
		}
	}

	log.Info("Users retrieved successfully", "count", len(result), "duration", time.Since(start))
	return result, nil
}
