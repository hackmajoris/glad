package database

import (
	"github.com/hackmajoris/glad/pkg/logger"

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
