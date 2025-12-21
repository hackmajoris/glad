package database

import (
	"sync"

	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/pkg/logger"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDBRepository implements all repository interfaces using DynamoDB single table design
// It provides implementations for:
// - UserRepository (user management)
// - MasterSkillRepository (master skills)
// - SkillRepository (user skills)
type DynamoDBRepository struct {
	client *dynamodb.DynamoDB
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository() *DynamoDBRepository {
	log := logger.WithComponent("database")
	log.Info("Initializing DynamoDB repository", "table", TableName)

	sess := session.Must(session.NewSession())
	repo := &DynamoDBRepository{
		client: dynamodb.New(sess),
	}

	log.Info("DynamoDB repository initialized successfully")
	return repo
}

// MockRepository implements UserRepository, SkillRepository, and MasterSkillRepository for testing
// This matches the DynamoDBRepository structure with unified implementation
type MockRepository struct {
	users        map[string]*models.User      // key: username
	skills       map[string]*models.UserSkill // key: "username#skillname"
	masterSkills map[string]*models.Skill     // key: skill_id
	mutex        sync.RWMutex
}

// NewMockRepository creates a new unified mock repository
func NewMockRepository() *MockRepository {
	log := logger.WithComponent("database")
	log.Info("Initializing unified Mock repository for local development")

	repo := &MockRepository{
		users:        make(map[string]*models.User),
		skills:       make(map[string]*models.UserSkill),
		masterSkills: make(map[string]*models.Skill),
	}

	log.Info("Unified Mock repository initialized successfully")
	return repo
}
