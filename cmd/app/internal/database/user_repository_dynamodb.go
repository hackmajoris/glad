package database

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// CreateUser inserts a new user into DynamoDB
func (r *DynamoDBRepository) CreateUser(user *models.User) error {
	log := logger.WithComponent("database").With("operation", "CreateUser", "username", user.Username)
	start := time.Now()

	log.Debug("Starting user creation")

	// Ensure keys are set
	user.SetKeys()

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Error("Failed to marshal user data", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(entity_id)"),
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

	entityID := models.BuildUserEntityID(username)
	log.Info("Attempting to retrieve user", "entity_id", entityID, "table", TableName)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"entity_id": {S: aws.String(entityID)},
		},
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		log.Error("Failed to get user from DynamoDB", "error", err.Error(), "entity_id", entityID, "duration", time.Since(start))
		return nil, err
	}

	if result.Item == nil {
		log.Info("User not found in DynamoDB", "entity_id", entityID, "duration", time.Since(start))
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

	entityID := models.BuildUserEntityID(username)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"entity_id": {S: aws.String(entityID)},
		},
		ProjectionExpression: aws.String("entity_id"),
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

	// Ensure keys are set
	user.SetKeys()
	user.UpdatedAt = time.Now()

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		log.Error("Failed to marshal user data for update", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(entity_id)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to update user in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("User updated successfully", "duration", time.Since(start))
	return nil
}

// ListUsers retrieves all users from DynamoDB using Query on EntityType
func (r *DynamoDBRepository) ListUsers() ([]*models.User, error) {
	log := logger.WithComponent("database").With("operation", "ListUsers")
	start := time.Now()

	log.Debug("Starting users list retrieval")

	// Use Scan with filter for EntityType = "User"
	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(GSIByEntityType),
		KeyConditionExpression: aws.String("EntityType = :entityType"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":entityType": {S: aws.String("User")},
		},
	}

	result, err := r.client.Query(input)
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
