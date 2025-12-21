package models

import (
	"errors"
	"time"

	apperrors "github.com/hackmajoris/glad/pkg/errors"
)

// Skill represents a master skill entity in the system
// This is the authoritative source for skill metadata
// UserSkills reference skills via skill_id and denormalize name/category
type Skill struct {
	// Business attributes
	SkillID     string    `json:"skill_id" dynamodbav:"skill_id"`    // Immutable ID (e.g., "python")
	SkillName   string    `json:"skill_name" dynamodbav:"SkillName"` // Display name (e.g., "Python")
	Description string    `json:"description" dynamodbav:"Description"`
	Category    string    `json:"category" dynamodbav:"Category"` // e.g., "Programming", "Cloud", "DevOps"
	Tags        []string  `json:"tags,omitempty" dynamodbav:"Tags,omitempty"`
	CreatedAt   time.Time `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time `json:"updated_at" dynamodbav:"UpdatedAt"`

	// DynamoDB attributes
	EntityID   string `json:"-" dynamodbav:"entity_id"`
	EntityType string `json:"entity_type" dynamodbav:"EntityType"`
}

// NewSkill creates a new master Skill
// skillID must be lowercase alphanumeric with dashes only (e.g., "python", "aws-lambda", "react-js")
// skillName is the display name (e.g., "Python", "AWS Lambda", "React.js")
// category should be a valid category (e.g., "Programming", "Cloud", "DevOps", "Database")
func NewSkill(skillID, skillName, description, category string, tags []string) (*Skill, error) {
	if skillID == "" || skillName == "" || category == "" {
		return nil, apperrors.ErrRequiredField
	}

	if !isValidSkillID(skillID) {
		return nil, errors.New("invalid skill_id: must be lowercase alphanumeric with dashes, max 50 chars")
	}

	if len(skillName) < 2 || len(skillName) > 100 {
		return nil, errors.New("invalid skill_name: must be between 2 and 100 characters")
	}

	if !isValidCategory(category) {
		return nil, errors.New("invalid category: must be one of Programming, Cloud, DevOps, Database, Frontend, Backend, Mobile, Data, Security, Other")
	}

	now := time.Now()
	skill := &Skill{
		SkillID:     skillID,
		SkillName:   skillName,
		Description: description,
		Category:    category,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	skill.SetKeys()
	return skill, nil
}

// isValidSkillID validates that a skill ID follows the required format:
// - lowercase letters (a-z)
// - numbers (0-9)
// - dashes (-)
// - length between 1 and 50 characters
func isValidSkillID(id string) bool {
	if id == "" || len(id) > 50 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// validCategories defines the allowed skill categories
var validCategories = map[string]bool{
	"Programming": true,
	"Cloud":       true,
	"DevOps":      true,
	"Database":    true,
	"Frontend":    true,
	"Backend":     true,
	"Mobile":      true,
	"Data":        true,
	"Security":    true,
	"Other":       true,
}

// isValidCategory checks if the category is in the allowed list
func isValidCategory(category string) bool {
	return validCategories[category]
}

// SetKeys configures the entity_id for DynamoDB
func (s *Skill) SetKeys() {
	s.EntityID = BuildMasterSkillEntityID(s.SkillID)
	s.EntityType = "Skill"
}

// UpdateMetadata updates skill display name, description, and category
// Note: This requires syncing all UserSkills that reference this skill
func (s *Skill) UpdateMetadata(skillName, description, category string) {
	if skillName != "" {
		s.SkillName = skillName
	}
	if description != "" {
		s.Description = description
	}
	if category != "" {
		s.Category = category
	}
	s.UpdatedAt = time.Now()
}

// UpdateTags updates the skill tags
func (s *Skill) UpdateTags(tags []string) {
	s.Tags = tags
	s.UpdatedAt = time.Now()
}
