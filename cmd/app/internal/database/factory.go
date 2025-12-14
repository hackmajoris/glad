package database

import (
	"os"

	"github.com/hackmajoris/glad/pkg/config"
	"github.com/hackmajoris/glad/pkg/logger"
)

// Repository combines all repository interfaces for unified access.
// Both DynamoDBRepository and MockRepository implement this interface.
type Repository interface {
	UserRepository
	SkillRepository
	MasterSkillRepository
}

// NewRepository creates the appropriate repository implementation based on configuration
func NewRepository(cfg *config.Config) Repository {
	log := logger.WithComponent("database")

	// Determine if we should use mock or real DynamoDB
	if shouldUseMockRepository(cfg) {
		log.Info("Creating Mock repository for development/testing")
		return NewMockRepository()
	}

	log.Info("Creating DynamoDB repository for production/Lambda")
	return NewDynamoDBRepository()
}

// shouldUseMockRepository determines if we should use mock repository
func shouldUseMockRepository(cfg *config.Config) bool {
	// 1. If AWS_LAMBDA_FUNCTION_NAME exists, we're definitely in Lambda - use DynamoDB
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return false
	}

	// 2. If ENVIRONMENT is explicitly set to production, use DynamoDB
	if os.Getenv("ENVIRONMENT") == "production" {
		return false
	}

	// 3. If LocalServer environment is development or test, use mock
	if cfg.IsDevelopment() {
		return true
	}

	// 4. If DB_MOCK is explicitly set to true, use mock (useful for testing)
	if os.Getenv("DB_MOCK") == "true" {
		return true
	}

	// 5. Default to DynamoDB for production
	return false
}
