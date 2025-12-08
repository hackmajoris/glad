package middleware

import (
	"testing"
	"time"

	"github.com/hackmajoris/glad/pkg/auth"
	"github.com/hackmajoris/glad/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

// MockUser implements the User interface for testing
type MockUser struct {
	Username string
}

func (m *MockUser) GetUsername() string {
	return m.Username
}

// testConfig creates a config for testing
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expiry: 24 * time.Hour,
		},
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	tokenService := auth.NewTokenService(testConfig())
	middleware := NewAuthMiddleware(tokenService)

	if middleware == nil {
		t.Error("Expected non-nil middleware")
		return
	}

	if middleware.tokenService != tokenService {
		t.Error("Expected middleware to store token service")
	}
}

func TestAuthMiddleware_ValidateJWT(t *testing.T) {
	tokenService := auth.NewTokenService(testConfig())
	middleware := NewAuthMiddleware(tokenService)
	user := &MockUser{Username: "testuser"}

	// Generate a valid token
	validToken, err := tokenService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Mock handler that should be called on success
	mockHandler := func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// Verify claims were injected
		claims, ok := request.RequestContext.Authorizer["claims"].(*auth.JWTClaims)
		if !ok {
			t.Error("Expected claims to be injected into request context")
		}
		if claims.Username != "testuser" {
			t.Errorf("Expected username %s, got %s", "testuser", claims.Username)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "success",
		}, nil
	}

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid token with Authorization header",
			headers: map[string]string{
				"Authorization": "Bearer " + validToken,
			},
			expectedStatus: 200,
			expectedBody:   "success",
		},
		{
			name: "valid token with lowercase authorization header",
			headers: map[string]string{
				"authorization": "Bearer " + validToken,
			},
			expectedStatus: 200,
			expectedBody:   "success",
		},
		{
			name:           "missing authorization header",
			headers:        map[string]string{},
			expectedStatus: 401,
			expectedBody:   `{"error": "Missing authorization token"}`,
		},
		{
			name: "invalid authorization header format",
			headers: map[string]string{
				"Authorization": "InvalidFormat " + validToken,
			},
			expectedStatus: 401,
			expectedBody:   `{"error": "Missing authorization token"}`,
		},
		{
			name: "invalid token",
			headers: map[string]string{
				"Authorization": "Bearer invalid.token.here",
			},
			expectedStatus: 401,
			expectedBody:   `{"error": "Invalid or expired token"}`,
		},
		{
			name: "empty token",
			headers: map[string]string{
				"Authorization": "Bearer ",
			},
			expectedStatus: 401,
			expectedBody:   `{"error": "Missing authorization token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := events.APIGatewayProxyRequest{
				Headers: tt.headers,
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: make(map[string]interface{}),
				},
			}

			protectedHandler := middleware.ValidateJWT(mockHandler)
			response, err := protectedHandler(request)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if response.Body != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, response.Body)
			}

			// Check Content-Type header
			if response.Headers["Content-Type"] != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", response.Headers["Content-Type"])
			}
		})
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name: "valid Authorization header",
			headers: map[string]string{
				"Authorization": "Bearer abc123",
			},
			expected: "abc123",
		},
		{
			name: "valid lowercase authorization header",
			headers: map[string]string{
				"authorization": "Bearer xyz789",
			},
			expected: "xyz789",
		},
		{
			name:     "no authorization header",
			headers:  map[string]string{},
			expected: "",
		},
		{
			name: "invalid format - no Bearer",
			headers: map[string]string{
				"Authorization": "Basic abc123",
			},
			expected: "",
		},
		{
			name: "invalid format - only Bearer",
			headers: map[string]string{
				"Authorization": "Bearer",
			},
			expected: "",
		},
		{
			name: "token with spaces",
			headers: map[string]string{
				"Authorization": "Bearer token with spaces",
			},
			expected: "token with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTokenFromHeader(tt.headers)
			if result != tt.expected {
				t.Errorf("extractTokenFromHeader() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnauthorizedResponse(t *testing.T) {
	message := "Test error message"
	response := unauthorizedResponse(message)

	if response.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", response.StatusCode)
	}

	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", response.Headers["Content-Type"])
	}

	expectedBody := `{"error": "Test error message"}`
	if response.Body != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, response.Body)
	}
}
