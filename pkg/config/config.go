package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	JWT         JWTConfig
	Database    DatabaseConfig
	LocalServer ServerConfig
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret     string
	Expiry     time.Duration
	SigningAlg string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	TableName string
	Region    string
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Environment string
	Port        int
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "default-secret-key"),
			Expiry:     getDurationEnv("JWT_EXPIRY", 24*time.Hour),
			SigningAlg: getEnv("JWT_SIGNING_ALG", "HS256"),
		},
		Database: DatabaseConfig{
			TableName: getEnv("DYNAMODB_TABLE", "users"),
			Region:    getEnv("AWS_REGION", "us-east-1"),
		},

		// local testing only
		LocalServer: ServerConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			Port:        getIntEnv("PORT", 8080),
		},
	}
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.LocalServer.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.LocalServer.Environment == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
