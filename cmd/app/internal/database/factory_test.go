package database

import (
	"os"
	"testing"
	"time"

	"github.com/hackmajoris/glad/pkg/config"
)

func TestNewRepository_EnvironmentDetection(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		configEnv   string
		expectMock  bool
		description string
	}{
		{
			name: "Lambda environment should use DynamoDB",
			setupEnv: func() {
				err := os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "glad-api")
				if err != nil {
					return
				}
			},
			cleanupEnv: func() {
				err := os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
				if err != nil {
					return
				}
			},
			configEnv:   "development",
			expectMock:  false,
			description: "When AWS_LAMBDA_FUNCTION_NAME is set, should use DynamoDB regardless of config",
		},
		{
			name: "Explicit production environment should use DynamoDB",
			setupEnv: func() {
				err := os.Setenv("ENVIRONMENT", "production")
				if err != nil {
					return
				}
			},
			cleanupEnv: func() {
				err := os.Unsetenv("ENVIRONMENT")
				if err != nil {
					return
				}
			},
			configEnv:   "development",
			expectMock:  false,
			description: "When ENVIRONMENT=production, should use DynamoDB",
		},
		{
			name: "Development config should use Mock",
			setupEnv: func() {
				// No special env vars set
			},
			cleanupEnv: func() {
				// Nothing to clean
			},
			configEnv:   "development",
			expectMock:  true,
			description: "When LocalServer.Environment=development, should use Mock",
		},
		{
			name: "Explicit DB_MOCK=true should use Mock",
			setupEnv: func() {
				err := os.Setenv("DB_MOCK", "true")
				if err != nil {
					return
				}
			},
			cleanupEnv: func() {
				err := os.Unsetenv("DB_MOCK")
				if err != nil {
					return
				}
			},
			configEnv:   "production",
			expectMock:  true,
			description: "When DB_MOCK=true, should use Mock even in production config",
		},
		{
			name: "Default production should use DynamoDB",
			setupEnv: func() {
				// No special env vars set
			},
			cleanupEnv: func() {
				// Nothing to clean
			},
			configEnv:   "production",
			expectMock:  false,
			description: "When LocalServer.Environment=production and no overrides, should use DynamoDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Create config with specified environment
			cfg := &config.Config{
				JWT: config.JWTConfig{
					Secret: "test-secret",
					Expiry: 24 * time.Hour,
				},
				LocalServer: config.ServerConfig{
					Environment: tt.configEnv,
					Port:        8080,
				},
			}

			// Test the factory
			repo := NewRepository(cfg)

			// Check if we got the expected repository type
			_, isMock := repo.(*MockRepository)
			_, isDynamoDB := repo.(*DynamoDBRepository)

			if tt.expectMock {
				if !isMock {
					t.Errorf("Expected MockRepository, but got %T. %s", repo, tt.description)
				}
				if isDynamoDB {
					t.Errorf("Got DynamoDBRepository when expecting MockRepository. %s", tt.description)
				}
			} else {
				if isMock {
					t.Errorf("Got MockRepository when expecting DynamoDBRepository. %s", tt.description)
				}
				if !isDynamoDB {
					t.Errorf("Expected DynamoDBRepository, but got %T. %s", repo, tt.description)
				}
			}
		})
	}
}

func TestShouldUseMockRepository(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		configEnv   string
		expected    bool
		description string
	}{
		{
			name: "AWS Lambda function name present",
			setupEnv: func() {
				err := os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "my-function")
				if err != nil {
					return
				}
			},
			cleanupEnv: func() {
				err := os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
				if err != nil {
					return
				}
			},
			configEnv:   "development",
			expected:    false,
			description: "Should return false when AWS_LAMBDA_FUNCTION_NAME is set",
		},
		{
			name:        "Development config",
			setupEnv:    func() {},
			cleanupEnv:  func() {},
			configEnv:   "development",
			expected:    true,
			description: "Should return true for development config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer tt.cleanupEnv()

			cfg := &config.Config{
				LocalServer: config.ServerConfig{
					Environment: tt.configEnv,
				},
			}

			// Test
			result := shouldUseMockRepository(cfg)

			// Verify
			if result != tt.expected {
				t.Errorf("Expected %v, got %v. %s", tt.expected, result, tt.description)
			}
		})
	}
}
