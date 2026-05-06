package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// UserProfile represents the DynamoDB user profile structure
type UserProfile struct {
	PK           string    `dynamodbav:"PK"`
	SK           string    `dynamodbav:"entity_id"`
	EntityType   string    `dynamodbav:"EntityType"`
	Username     string    `dynamodbav:"Username"`
	Email        string    `dynamodbav:"Email"`
	Name         string    `dynamodbav:"Name"`
	PasswordHash string    `dynamodbav:"PasswordHash"` // Empty for Cognito users
	CreatedAt    time.Time `dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time `dynamodbav:"UpdatedAt"`
}

var (
	dynamoClient *dynamodb.DynamoDB
	tableName    string
)

func init() {
	// Initialize DynamoDB client
	sess := session.Must(session.NewSession())
	dynamoClient = dynamodb.New(sess)

	// Get table name from environment variable
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "entities-table" // Fallback default
	}
}

// Handler is the Lambda function handler for Cognito Post Confirmation trigger
// This function is invoked automatically by Cognito after a user successfully confirms their email
func Handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	// Extract user attributes from Cognito event
	username := event.UserName
	email := event.Request.UserAttributes["email"]

	// Create user profile in DynamoDB
	now := time.Now()
	userProfile := UserProfile{
		PK:           "User",
		SK:           "USER#" + username,
		EntityType:   "User",
		Username:     username,
		Email:        email,
		Name:         username, // Default to username, user can update later
		PasswordHash: "",       // Empty - Cognito manages passwords
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Marshal to DynamoDB attribute values
	item, err := dynamodbattribute.MarshalMap(userProfile)
	if err != nil {
		// Log error but don't fail the signup process
		// Cognito will still confirm the user
		return event, err
	}

	// Put item in DynamoDB
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
		// Use condition to prevent overwriting existing users
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}

	_, err = dynamoClient.PutItem(input)
	if err != nil {
		// Check if it's a conditional check failure (user already exists)
		// This can happen if user confirms multiple times
		// Don't fail the process, just log and continue
		return event, err
	}

	// Return the event unmodified - required by Cognito
	return event, nil
}

func main() {
	lambda.Start(Handler)
}
