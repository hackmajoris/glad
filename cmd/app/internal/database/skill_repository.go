package database

import (
	"fmt"
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// ============================================================================
// SKILL REPOSITORY METHODS
// ============================================================================

type SkillRepository interface {
	CreateSkill(skill *models.UserSkill) error
	GetSkill(username, skillName string) (*models.UserSkill, error)
	UpdateSkill(skill *models.UserSkill) error
	DeleteSkill(username, skillName string) error
	ListSkillsForUser(username string) ([]*models.UserSkill, error)
	ListUsersBySkill(skillName string) ([]*models.UserSkill, error)
	ListUsersBySkillAndLevel(skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error)
}

// CreateSkill inserts a new user skill into DynamoDB
func (r *DynamoDBRepository) CreateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "CreateSkill", "username", skill.Username, "skill", skill.SkillName)
	start := time.Now()

	log.Debug("Starting skill creation")

	// Ensure keys are set
	skill.SetKeys()

	item, err := dynamodbattribute.MarshalMap(skill)
	if err != nil {
		log.Error("Failed to marshal skill data", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to create skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillAlreadyExists
	}

	log.Info("Skill created successfully", "duration", time.Since(start))
	return nil
}

// GetSkill retrieves a specific skill for a user
func (r *DynamoDBRepository) GetSkill(username, skillName string) (*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "GetSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Debug("Starting skill retrieval")

	pk := fmt.Sprintf("USER#%s", username)
	sk := fmt.Sprintf("SKILL#%s", skillName)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(pk)},
			"SK": {S: aws.String(sk)},
		},
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		log.Error("Failed to get skill from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	if result.Item == nil {
		log.Debug("Skill not found", "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	var skill models.UserSkill
	err = dynamodbattribute.UnmarshalMap(result.Item, &skill)
	if err != nil {
		log.Error("Failed to unmarshal skill data", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Debug("Skill retrieved successfully", "duration", time.Since(start))
	return &skill, nil
}

// UpdateSkill updates an existing skill
func (r *DynamoDBRepository) UpdateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "UpdateSkill", "username", skill.Username, "skill", skill.SkillName)
	start := time.Now()

	log.Debug("Starting skill update")

	// Ensure keys are set
	skill.SetKeys()
	skill.UpdatedAt = time.Now()

	item, err := dynamodbattribute.MarshalMap(skill)
	if err != nil {
		log.Error("Failed to marshal skill data for update", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to update skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	log.Info("Skill updated successfully", "duration", time.Since(start))
	return nil
}

// DeleteSkill removes a skill from a user
func (r *DynamoDBRepository) DeleteSkill(username, skillName string) error {
	log := logger.WithComponent("database").With("operation", "DeleteSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Debug("Starting skill deletion")

	pk := fmt.Sprintf("USER#%s", username)
	sk := fmt.Sprintf("SKILL#%s", skillName)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(pk)},
			"SK": {S: aws.String(sk)},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}

	_, err := r.client.DeleteItem(input)
	if err != nil {
		log.Error("Failed to delete skill from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	log.Info("Skill deleted successfully", "duration", time.Since(start))
	return nil
}

// ListSkillsForUser retrieves all skills for a specific user (item collection query)
func (r *DynamoDBRepository) ListSkillsForUser(username string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListSkillsForUser", "username", username)
	start := time.Now()

	log.Debug("Starting skills list retrieval for user")

	pk := fmt.Sprintf("USER#%s", username)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk":        {S: aws.String(pk)},
			":sk_prefix": {S: aws.String("SKILL#")},
		},
	}

	result, err := r.client.Query(input)
	if err != nil {
		log.Error("Failed to query skills for user", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	var skills []*models.UserSkill
	for i, item := range result.Items {
		var skill models.UserSkill
		if err := dynamodbattribute.UnmarshalMap(item, &skill); err != nil {
			log.Error("Failed to unmarshal skill data", "error", err.Error(), "item_index", i, "duration", time.Since(start))
			continue
		}
		skills = append(skills, &skill)
	}

	log.Info("Skills retrieved successfully", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkill retrieves all users who have a specific skill (GSI1 query)
func (r *DynamoDBRepository) ListUsersBySkill(skillName string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkill", "skill", skillName)
	start := time.Now()

	log.Debug("Starting users list retrieval by skill")

	gsi1pk := fmt.Sprintf("SKILL#%s", skillName)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(GSI1Name),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {S: aws.String(gsi1pk)},
		},
	}

	result, err := r.client.Query(input)
	if err != nil {
		log.Error("Failed to query users by skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	var skills []*models.UserSkill
	for i, item := range result.Items {
		var skill models.UserSkill
		if err := dynamodbattribute.UnmarshalMap(item, &skill); err != nil {
			log.Error("Failed to unmarshal skill data", "error", err.Error(), "item_index", i, "duration", time.Since(start))
			continue
		}
		skills = append(skills, &skill)
	}

	log.Info("Users with skill retrieved successfully", "skill", skillName, "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkillAndLevel retrieves users with a specific skill at a specific proficiency level (GSI1 query with sort key filter)
func (r *DynamoDBRepository) ListUsersBySkillAndLevel(skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkillAndLevel", "skill", skillName, "level", proficiencyLevel)
	start := time.Now()

	log.Debug("Starting users list retrieval by skill and level")

	gsi1pk := fmt.Sprintf("SKILL#%s", skillName)
	gsi1skPrefix := fmt.Sprintf("LEVEL#%s#", proficiencyLevel)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(GSI1Name),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk AND begins_with(GSI1SK, :gsi1sk_prefix)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk":        {S: aws.String(gsi1pk)},
			":gsi1sk_prefix": {S: aws.String(gsi1skPrefix)},
		},
	}

	result, err := r.client.Query(input)
	if err != nil {
		log.Error("Failed to query users by skill and level", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	var skills []*models.UserSkill
	for i, item := range result.Items {
		var skill models.UserSkill
		if err := dynamodbattribute.UnmarshalMap(item, &skill); err != nil {
			log.Error("Failed to unmarshal skill data", "error", err.Error(), "item_index", i, "duration", time.Since(start))
			continue
		}
		skills = append(skills, &skill)
	}

	log.Info("Users with skill and level retrieved successfully", "skill", skillName, "level", proficiencyLevel, "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
