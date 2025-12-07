package database

import (
	"github.com/hackmajoris/glad/pkg/logger"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	// TableName is the single table for all entities
	TableName = "glad-entities"

	// GSI1Name GSI names
	GSI1Name = "GSI1"
)

// DynamoDBRepository implements Repository using DynamoDB single table design
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
