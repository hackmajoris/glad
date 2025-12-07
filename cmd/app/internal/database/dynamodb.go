package database

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	UsersTableName = "users"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUser(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	UserExists(username string) (bool, error)
	ListUsers() ([]*models.User, error)
}

// DynamoDBRepository implements UserRepository using DynamoDB
type DynamoDBRepository struct {
	client *dynamodb.DynamoDB
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository() *DynamoDBRepository {
	log := logger.WithComponent("database")
	log.Info("Initializing DynamoDB repository", "table", UsersTableName)

	sess := session.Must(session.NewSession())
	repo := &DynamoDBRepository{
		client: dynamodb.New(sess),
	}

	log.Info("DynamoDB repository initialized successfully")
	return repo
}

// CreateUser inserts a new user into DynamoDB
func (r *DynamoDBRepository) CreateUser(user *models.User) error {
	log := logger.WithComponent("database").With("operation", "CreateUser", "username", user.Username)
	start := time.Now()

	log.Debug("Starting user creation")

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Error("Failed to marshal user data", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(UsersTableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(username)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to create user in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("User created successfully", "duration", time.Since(start))
	return nil
}

// GetUser retrieves a user by username from DynamoDB
func (r *DynamoDBRepository) GetUser(username string) (*models.User, error) {
	log := logger.WithComponent("database").With("operation", "GetUser", "username", username)
	start := time.Now()

	log.Debug("Starting user retrieval")

	input := &dynamodb.GetItemInput{
		TableName: aws.String(UsersTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		log.Error("Failed to get user from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	if result.Item == nil {
		log.Debug("User not found", "duration", time.Since(start))
		return nil, apperrors.ErrUserNotFound
	}

	var user models.User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		log.Error("Failed to unmarshal user data", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Debug("User retrieved successfully", "duration", time.Since(start))
	return &user, nil
}

// UserExists checks if a user exists in DynamoDB
func (r *DynamoDBRepository) UserExists(username string) (bool, error) {
	log := logger.WithComponent("database").With("operation", "UserExists", "username", username)
	start := time.Now()

	log.Debug("Checking if user exists")

	input := &dynamodb.GetItemInput{
		TableName: aws.String(UsersTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
		ProjectionExpression: aws.String("username"),
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		log.Error("Failed to check user existence", "error", err.Error(), "duration", time.Since(start))
		return false, err
	}

	exists := result.Item != nil
	log.Debug("User existence check completed", "exists", exists, "duration", time.Since(start))
	return exists, nil
}

// UpdateUser updates an existing user in DynamoDB
func (r *DynamoDBRepository) UpdateUser(user *models.User) error {
	log := logger.WithComponent("database").With("operation", "UpdateUser", "username", user.Username)
	start := time.Now()

	log.Debug("Starting user update")

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Error("Failed to marshal user data for update", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(UsersTableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(username)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to update user in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("User updated successfully", "duration", time.Since(start))
	return nil
}

// ListUsers retrieves all users from DynamoDB
func (r *DynamoDBRepository) ListUsers() ([]*models.User, error) {
	log := logger.WithComponent("database").With("operation", "ListUsers")
	start := time.Now()

	log.Debug("Starting users list retrieval")

	input := &dynamodb.ScanInput{
		TableName: aws.String(UsersTableName),
	}

	result, err := r.client.Scan(input)
	if err != nil {
		log.Error("Failed to scan users table", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	var users []*models.User
	for i, item := range result.Items {
		var user models.User
		if err := dynamodbattribute.UnmarshalMap(item, &user); err != nil {
			log.Error("Failed to unmarshal user data", "error", err.Error(), "item_index", i, "duration", time.Since(start))
			return nil, err
		}
		users = append(users, &user)
	}

	log.Info("Users retrieved successfully", "count", len(users), "scanned_count", *result.ScannedCount, "duration", time.Since(start))
	return users, nil
}
