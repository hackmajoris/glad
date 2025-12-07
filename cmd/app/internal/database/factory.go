package database

import (
	"github.com/hackmajoris/glad/pkg/config"
	"github.com/hackmajoris/glad/pkg/logger"
)

// Repository combines all repository interfaces for unified access
type Repository interface {
	UserRepository
	SkillRepository
}

// NewRepository creates the appropriate repository implementation based on configuration
func NewRepository(cfg *config.Config) Repository {
	log := logger.WithComponent("database")

	if cfg.LocalServer.Environment == "development" || cfg.LocalServer.Environment == "test" {
		log.Info("Creating Mock repository for development/testing")
		return NewMockRepository()
	}

	log.Info("Creating DynamoDB repository for production")
	return NewDynamoDBRepository()
}
