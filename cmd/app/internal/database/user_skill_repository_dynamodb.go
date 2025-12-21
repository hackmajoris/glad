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

// CreateSkill inserts a new user skill into DynamoDB
func (r *DynamoDBRepository) CreateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "CreateSkill", "username", skill.Username, "skill_id", skill.SkillID)
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
		ConditionExpression: aws.String("attribute_not_exists(entity_id)"),
	}
	_, err = r.client.PutItem(input)

	if err != nil {
		log.Error("Failed to create skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("Skill created successfully", "duration", time.Since(start))

	return nil
}

// GetSkill retrieves a specific skill for a user by skill_id
func (r *DynamoDBRepository) GetSkill(username, skillID string) (*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "GetSkill", "username", username, "skill_id", skillID)
	start := time.Now()

	log.Debug("Starting skill retrieval")

	entityID := BuildUserSkillEntityID(username, skillID)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"EntityType": {S: aws.String("UserSkill")},
			"entity_id":  {S: aws.String(entityID)},
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
	log := logger.WithComponent("database").With("operation", "UpdateSkill", "username", skill.Username, "skill_id", skill.SkillID)
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
		ConditionExpression: aws.String("attribute_exists(entity_id)"),
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		log.Error("Failed to update skill in DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("Skill updated successfully", "duration", time.Since(start))
	return nil
}

// DeleteSkill removes a skill from a user
func (r *DynamoDBRepository) DeleteSkill(username, skillID string) error {
	log := logger.WithComponent("database").With("operation", "DeleteSkill", "username", username, "skill_id", skillID)
	start := time.Now()

	log.Debug("Starting skill deletion")

	entityID := BuildUserSkillEntityID(username, skillID)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"EntityType": {S: aws.String("UserSkill")},
			"entity_id":  {S: aws.String(entityID)},
		},
		ConditionExpression: aws.String("attribute_exists(entity_id)"),
	}

	_, err := r.client.DeleteItem(input)
	if err != nil {
		log.Error("Failed to delete skill from DynamoDB", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("Skill deleted successfully", "duration", time.Since(start))
	return nil
}

// ListSkillsForUser retrieves all skills for a specific user using GSI ByUser
func (r *DynamoDBRepository) ListSkillsForUser(username string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListSkillsForUser", "username", username)
	start := time.Now()

	log.Debug("Starting skills list retrieval for user")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		KeyConditionExpression: aws.String("EntityType = :entityType AND begins_with(entity_id, :userPrefix)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":entityType": {S: aws.String("UserSkill")},
			":userPrefix": {S: aws.String("USERSKILL#" + username + "#")},
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

// ListUsersBySkill retrieves all users who have a specific skill using GSI BySkill
// GSI BySkill structure: PK=Category, SK=SkillName+ProficiencyLevel+YearsOfExperience+Username
func (r *DynamoDBRepository) ListUsersBySkill(category, skillName string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkill", "category", category, "skill", skillName)
	start := time.Now()

	log.Debug("Starting users list retrieval by skill")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(GSIBySkill),
		KeyConditionExpression: aws.String("Category = :category AND SkillName = :skillName"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":category":  {S: aws.String(category)},
			":skillName": {S: aws.String(skillName)},
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

	log.Info("Users with skill retrieved successfully", "category", category, "skill", skillName, "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkillAndLevel retrieves users with a specific skill at a specific proficiency level
// GSI BySkill structure: PK=Category, SK=SkillName+ProficiencyLevel+YearsOfExperience+Username
// Uses composite sort key matching: Category + SkillName + ProficiencyLevel (left-to-right)
func (r *DynamoDBRepository) ListUsersBySkillAndLevel(category, skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkillAndLevel", "category", category, "skill", skillName, "level", proficiencyLevel)
	start := time.Now()

	log.Debug("Starting users list retrieval by skill and level")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String(GSIBySkill),
		KeyConditionExpression: aws.String("Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":category":  {S: aws.String(category)},
			":skillName": {S: aws.String(skillName)},
			":level":     {S: aws.String(string(proficiencyLevel))},
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

	log.Info("Users with skill and level retrieved successfully", "category", category, "skill", skillName, "level", proficiencyLevel, "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
