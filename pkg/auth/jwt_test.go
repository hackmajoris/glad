package auth

import (
	"os"
	"testing"
	"time"

	"github.com/hackmajoris/glad-stack/pkg/config"

	"github.com/golang-jwt/jwt/v5"
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

func TestNewTokenService(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "with environment variable",
			envValue: "test-secret-key",
			expected: "test-secret-key",
		},
		{
			name:     "without environment variable",
			envValue: "",
			expected: "default-secret-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			original := os.Getenv("JWT_SECRET")
			defer func(key, value string) {
				err := os.Setenv(key, value)
				if err != nil {

				}
			}("JWT_SECRET", original)

			// Set test env value
			if tt.envValue != "" {
				err := os.Setenv("JWT_SECRET", tt.envValue)
				if err != nil {
					return
				}
			} else {
				err := os.Unsetenv("JWT_SECRET")
				if err != nil {
					return
				}
			}

			cfg := config.Load()
			ts := NewTokenService(cfg)
			if string(ts.secretKey) != tt.expected {
				t.Errorf("Expected secret key %s, got %s", tt.expected, string(ts.secretKey))
			}
		})
	}
}

func TestTokenService_GenerateToken(t *testing.T) {
	ts := NewTokenService(testConfig())
	user := &MockUser{Username: "testuser"}

	token, err := ts.GenerateToken(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return ts.secretKey, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token is not valid")
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to cast claims to JWTClaims")
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username %s, got %s", "testuser", claims.Username)
	}

	if claims.Subject != "testuser" {
		t.Errorf("Expected subject %s, got %s", "testuser", claims.Subject)
	}

	// Verify expiration is set to 24 hours from now
	expectedExp := time.Now().Add(24 * time.Hour)
	actualExp := claims.ExpiresAt.Time
	if actualExp.Sub(expectedExp) > time.Minute || expectedExp.Sub(actualExp) > time.Minute {
		t.Errorf("Token expiration time is not approximately 24 hours from now")
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	ts := NewTokenService(testConfig())
	user := &MockUser{Username: "testuser"}

	// Generate a valid token
	validToken, err := ts.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.jwt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ts.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenService.ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims == nil {
					t.Error("Expected claims, got nil")
					return
				}
				if claims.Username != "testuser" {
					t.Errorf("Expected username %s, got %s", "testuser", claims.Username)
				}
			}
		})
	}
}

func TestTokenService_ValidateExpiredToken(t *testing.T) {
	ts := NewTokenService(testConfig())

	// Create an expired token
	claims := JWTClaims{
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "testuser",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ts.secretKey)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ts.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestTokenService_ValidateTokenWithWrongSecret(t *testing.T) {
	cfg1 := &config.Config{
		JWT: config.JWTConfig{
			Secret: "secret-one",
			Expiry: 24 * time.Hour,
		},
	}
	ts1 := NewTokenService(cfg1)
	user := &MockUser{Username: "testuser"}

	// Generate token with first service
	token, err := ts1.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	cfg2 := &config.Config{
		JWT: config.JWTConfig{
			Secret: "different-secret",
			Expiry: 24 * time.Hour,
		},
	}
	ts2 := NewTokenService(cfg2)

	_, err = ts2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret, got nil")
	}
}
