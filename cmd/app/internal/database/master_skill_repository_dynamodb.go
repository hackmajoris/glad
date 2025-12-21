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

// CreateMasterSkill inserts a new master skill
func (r *DynamoDBRepository) CreateMasterSkill(skill *models.Skill) error {
	log := logger.WithComponent("database").With("operation", "CreateMasterSkill", "skill_id", skill.SkillID)
	start := time.Now()

	log.Debug("Starting master skill creation")

	skill.SetKeys()

	item, err := dynamodbattribute.MarshalMap(skill)
	if err != nil {
		log.Error("Failed to marshal skill data", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(entity_id)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to create master skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillAlreadyExists
	}

	log.Info("Master skill created successfully", "duration", time.Since(start))
	return nil
}

// GetMasterSkill retrieves a master skill by ID
func (r *DynamoDBRepository) GetMasterSkill(skillID string) (*models.Skill, error) {
	log := logger.WithComponent("database").With("operation", "GetMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Debug("Starting master skill retrieval")

	entityID := BuildMasterSkillEntityID(skillID)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"EntityType": {S: aws.String("Skill")},
			"entity_id":  {S: aws.String(entityID)},
		},
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		log.Error("Failed to get master skill from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	if result.Item == nil {
		log.Debug("Master skill not found", "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	var skill models.Skill
	err = dynamodbattribute.UnmarshalMap(result.Item, &skill)
	if err != nil {
		log.Error("Failed to unmarshal skill data", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Debug("Master skill retrieved successfully", "duration", time.Since(start))
	return &skill, nil
}

// UpdateMasterSkill updates an existing master skill
func (r *DynamoDBRepository) UpdateMasterSkill(skill *models.Skill) error {
	log := logger.WithComponent("database").With("operation", "UpdateMasterSkill", "skill_id", skill.SkillID)
	start := time.Now()

	log.Debug("Starting master skill update")

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
		ConditionExpression: aws.String("attribute_exists(entity_id)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to update master skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	log.Info("Master skill updated successfully", "duration", time.Since(start))
	return nil
}

// DeleteMasterSkill removes a master skill
func (r *DynamoDBRepository) DeleteMasterSkill(skillID string) error {
	log := logger.WithComponent("database").With("operation", "DeleteMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Debug("Starting master skill deletion")

	entityID := BuildMasterSkillEntityID(skillID)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"EntityType": {S: aws.String("Skill")},
			"entity_id":  {S: aws.String(entityID)},
		},
		ConditionExpression: aws.String("attribute_exists(entity_id)"),
	}

	_, err := r.client.DeleteItem(input)
	if err != nil {
		log.Error("Failed to delete master skill from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	log.Info("Master skill deleted successfully", "duration", time.Since(start))
	return nil
}

// ListMasterSkills retrieves all master skills
func (r *DynamoDBRepository) ListMasterSkills() ([]*models.Skill, error) {
	log := logger.WithComponent("database").With("operation", "ListMasterSkills")
	start := time.Now()

	log.Debug("Starting master skills list retrieval")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		KeyConditionExpression: aws.String("EntityType = :entityType"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":entityType": {S: aws.String("Skill")},
		},
	}

	result, err := r.client.Query(input)
	if err != nil {
		log.Error("Failed to query master skills", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	var skills []*models.Skill
	for i, item := range result.Items {
		var skill models.Skill
		if err := dynamodbattribute.UnmarshalMap(item, &skill); err != nil {
			log.Error("Failed to unmarshal skill data", "error", err.Error(), "item_index", i, "duration", time.Since(start))
			continue
		}
		skills = append(skills, &skill)
	}

	log.Info("Master skills retrieved successfully", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
