package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hackmajoris/glad-stack/cmd/glad/internal/database"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/dto"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/service"
	"github.com/hackmajoris/glad-stack/pkg/auth"
	"github.com/hackmajoris/glad-stack/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

// testConfig creates a config for testing
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expiry: 24 * time.Hour,
		},
	}
}

func TestHandler_GetCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(repo *database.MockRepository)
		claims         *auth.JWTClaims
		expectedStatus int
		validateBody   func(t *testing.T, body string)
	}{
		{
			name: "successful user retrieval",
			setupRepo: func(repo *database.MockRepository) {
				user, _ := models.NewUser("testuser", "Test User", "password123")
				user.CreatedAt = time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
				user.UpdatedAt = time.Date(2025, 1, 2, 15, 30, 0, 0, time.UTC)
				err := repo.CreateUser(user)
				if err != nil {
					return
				}
			},
			claims: &auth.JWTClaims{
				Username: "testuser",
			},
			expectedStatus: 200,
			validateBody: func(t *testing.T, body string) {
				var response dto.CurrentUserResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Username != "testuser" {
					t.Errorf("Expected username 'testuser', got '%s'", response.Username)
				}
				if response.Name != "Test User" {
					t.Errorf("Expected name 'Test User', got '%s'", response.Name)
				}
				if response.CreatedAt != "2025-01-01T10:00:00Z" {
					t.Errorf("Expected CreatedAt '2025-01-01T10:00:00Z', got '%s'", response.CreatedAt)
				}
				if response.UpdatedAt != "2025-01-02T15:30:00Z" {
					t.Errorf("Expected UpdatedAt '2025-01-02T15:30:00Z', got '%s'", response.UpdatedAt)
				}
			},
		},
		{
			name: "invalid token claims",
			setupRepo: func(repo *database.MockRepository) {
				// No setup needed
			},
			claims:         nil,
			expectedStatus: 401,
			validateBody: func(t *testing.T, body string) {
				var response dto.ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if response.Error != "Invalid token claims" {
					t.Errorf("Expected error 'Invalid token claims', got '%s'", response.Error)
				}
			},
		},
		{
			name: "user not found",
			setupRepo: func(repo *database.MockRepository) {
				// Don't create the user
			},
			claims: &auth.JWTClaims{
				Username: "nonexistent",
			},
			expectedStatus: 404,
			validateBody: func(t *testing.T, body string) {
				var response dto.ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if response.Error != "User not found" {
					t.Errorf("Expected error 'User not found', got '%s'", response.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create unified mock repository
			mockRepo := database.NewMockRepository()
			masterSkillsRepo := database.NewMockRepository()

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			// Create services with mock repository
			tokenService := auth.NewTokenService(testConfig())
			userService := service.NewUserService(mockRepo, tokenService)
			skillService := service.NewSkillService(mockRepo, masterSkillsRepo, mockRepo)

			// Create handler
			h := New(userService, skillService)

			// Create request
			request := events.APIGatewayProxyRequest{
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: make(map[string]interface{}),
				},
			}

			// Set claims if provided
			if tt.claims != nil {
				request.RequestContext.Authorizer["claims"] = tt.claims
			}

			// Call handler
			response, err := h.GetCurrentUser(request)

			// Verify no error from handler
			if err != nil {
				t.Fatalf("Handler returned unexpected error: %v", err)
			}

			// Verify status code
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			// Verify Content-Type header
			if response.Headers["Content-Type"] != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", response.Headers["Content-Type"])
			}

			// Validate response body
			if tt.validateBody != nil {
				tt.validateBody(t, response.Body)
			}
		})
	}
}

// TestHandler_GetCurrentUser_TimestampFormat verifies the timestamp format is ISO 8601
func TestHandler_GetCurrentUser_TimestampFormat(t *testing.T) {
	// Create unified mock repository
	mockRepo := database.NewMockRepository()

	// Create a user with specific timestamps
	user, _ := models.NewUser("testuser", "Test User", "password123")
	user.CreatedAt = time.Date(2025, 12, 7, 14, 30, 45, 0, time.FixedZone("EST", -5*3600))
	user.UpdatedAt = time.Date(2025, 12, 7, 16, 45, 30, 0, time.FixedZone("PST", -8*3600))
	err := mockRepo.CreateUser(user)
	if err != nil {
		return
	}

	tokenService := auth.NewTokenService(testConfig())
	userService := service.NewUserService(mockRepo, tokenService)
	mockRepository := database.NewMockRepository()
	masterSkillRepository := database.NewMockRepository()
	skillService := service.NewSkillService(mockRepository, masterSkillRepository, mockRepo)
	h := New(userService, skillService)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": &auth.JWTClaims{Username: "testuser"},
			},
		},
	}

	response, err := h.GetCurrentUser(request)
	if err != nil {
		t.Fatalf("Handler returned unexpected error: %v", err)
	}

	var result dto.CurrentUserResponse
	if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify ISO 8601 format (RFC3339)
	expectedCreatedAt := "2025-12-07T14:30:45-05:00"
	expectedUpdatedAt := "2025-12-07T16:45:30-08:00"

	if result.CreatedAt != expectedCreatedAt {
		t.Errorf("Expected CreatedAt '%s', got '%s'", expectedCreatedAt, result.CreatedAt)
	}

	if result.UpdatedAt != expectedUpdatedAt {
		t.Errorf("Expected UpdatedAt '%s', got '%s'", expectedUpdatedAt, result.UpdatedAt)
	}
}

// TestHandler_GetCurrentUser_DoesNotExposePassword verifies password hash is not included
func TestHandler_GetCurrentUser_DoesNotExposePassword(t *testing.T) {
	// Create mock repository and service
	mockRepo := database.NewMockRepository()

	user, _ := models.NewUser("testuser", "Test User", "password123")
	err := mockRepo.CreateUser(user)
	if err != nil {
		return
	}

	tokenService := auth.NewTokenService(testConfig())
	userService := service.NewUserService(mockRepo, tokenService)
	skillMockRepo := database.NewMockRepository()
	masterSkillMockRepo := database.NewMockRepository()
	skillService := service.NewSkillService(skillMockRepo, masterSkillMockRepo, mockRepo)
	h := New(userService, skillService)

	request := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				"claims": &auth.JWTClaims{Username: "testuser"},
			},
		},
	}

	response, err := h.GetCurrentUser(request)
	if err != nil {
		t.Fatalf("Handler returned unexpected error: %v", err)
	}

	// Parse as generic map to check for password fields
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Ensure password-related fields are not present
	sensitiveFields := []string{"password", "password_hash", "passwordHash", "PasswordHash"}
	for _, field := range sensitiveFields {
		if _, exists := result[field]; exists {
			t.Errorf("Response should not contain sensitive field '%s'", field)
		}
	}

	// Verify expected fields are present
	expectedFields := []string{"username", "name", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Response should contain field '%s'", field)
		}
	}
}
