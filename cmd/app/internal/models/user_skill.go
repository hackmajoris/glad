package models

import (
	"fmt"
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/pkg/errors"
)

// ProficiencyLevel represents the proficiency level for a skill
type ProficiencyLevel string

const (
	ProficiencyBeginner     ProficiencyLevel = "Beginner"
	ProficiencyIntermediate ProficiencyLevel = "Intermediate"
	ProficiencyAdvanced     ProficiencyLevel = "Advanced"
	ProficiencyExpert       ProficiencyLevel = "Expert"
)

// Valid proficiency levels
var validProficiencyLevels = map[ProficiencyLevel]bool{
	ProficiencyBeginner:     true,
	ProficiencyIntermediate: true,
	ProficiencyAdvanced:     true,
	ProficiencyExpert:       true,
}

// UserSkill represents a skill associated with a user (domain model)
// This entity uses single table design with the following key structure:
//   - PK: USER#<username>
//   - SK: SKILL#<skill_name>
//   - GSI1PK: SKILL#<skill_name>
//   - GSI1SK: LEVEL#<proficiency>#USER#<username>
type UserSkill struct {
	// Business attributes
	Username          string           `json:"username" dynamodbav:"Username"`
	SkillName         string           `json:"skill_name" dynamodbav:"SkillName"`
	ProficiencyLevel  ProficiencyLevel `json:"proficiency_level" dynamodbav:"ProficiencyLevel"`
	YearsOfExperience int              `json:"years_of_experience" dynamodbav:"YearsOfExperience"`
	Endorsements      int              `json:"endorsements" dynamodbav:"Endorsements"`
	LastUsedDate      string           `json:"last_used_date" dynamodbav:"LastUsedDate"` // ISO 8601 format
	Notes             string           `json:"notes,omitempty" dynamodbav:"Notes,omitempty"`
	CreatedAt         time.Time        `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt         time.Time        `json:"updated_at" dynamodbav:"UpdatedAt"`

	// DynamoDB system attributes for single table design
	PK         string `json:"-" dynamodbav:"PK"`
	SK         string `json:"-" dynamodbav:"SK"`
	EntityType string `json:"entity_type" dynamodbav:"EntityType"`

	// GSI1 attributes for cross-user skill queries
	GSI1PK string `json:"-" dynamodbav:"GSI1PK,omitempty"`
	GSI1SK string `json:"-" dynamodbav:"GSI1SK,omitempty"`
}

// NewUserSkill creates a new UserSkill with proper validation
func NewUserSkill(username, skillName string, proficiencyLevel ProficiencyLevel, yearsOfExperience int) (*UserSkill, error) {
	if username == "" {
		return nil, errors.ErrRequiredField
	}

	if skillName == "" {
		return nil, errors.ErrRequiredField
	}

	if !validProficiencyLevels[proficiencyLevel] {
		return nil, apperrors.ErrInvalidProficiencyLevel
	}

	if yearsOfExperience < 0 {
		return nil, apperrors.ErrInvalidYearsOfExperience
	}

	now := time.Now()
	skill := &UserSkill{
		Username:          username,
		SkillName:         skillName,
		ProficiencyLevel:  proficiencyLevel,
		YearsOfExperience: yearsOfExperience,
		Endorsements:      0,
		LastUsedDate:      now.Format("2006-01-02"), // ISO 8601 date format
		CreatedAt:         now,
		UpdatedAt:         now,
		EntityType:        "UserSkill",
	}

	// Set DynamoDB keys
	skill.SetKeys()

	return skill, nil
}

// SetKeys configures the PK, SK, and GSI keys for DynamoDB single table design
func (s *UserSkill) SetKeys() {
	// Base table keys: Item collection pattern
	// All skills for a user share the same PK
	s.PK = fmt.Sprintf("USER#%s", s.Username)
	s.SK = fmt.Sprintf("SKILL#%s", s.SkillName)

	// Entity type for filtering
	s.EntityType = "UserSkill"

	// GSI1 keys: For querying users by skill
	// Enables: "Find all users with skill X" or "Find all expert users with skill X"
	s.GSI1PK = fmt.Sprintf("SKILL#%s", s.SkillName)
	s.GSI1SK = fmt.Sprintf("LEVEL#%s#USER#%s", s.ProficiencyLevel, s.Username)
}

// UpdateProficiency updates the skill proficiency level
func (s *UserSkill) UpdateProficiency(level ProficiencyLevel) error {
	if !validProficiencyLevels[level] {
		return apperrors.ErrInvalidProficiencyLevel
	}

	s.ProficiencyLevel = level
	s.UpdatedAt = time.Now()

	// Update GSI keys to reflect new proficiency
	s.SetKeys()

	return nil
}

// UpdateYearsOfExperience updates the years of experience
func (s *UserSkill) UpdateYearsOfExperience(years int) error {
	if years < 0 {
		return apperrors.ErrInvalidYearsOfExperience
	}

	s.YearsOfExperience = years
	s.UpdatedAt = time.Now()

	return nil
}

// UpdateLastUsed updates the last used date to now
func (s *UserSkill) UpdateLastUsed() {
	s.LastUsedDate = time.Now().Format("2006-01-02")
	s.UpdatedAt = time.Now()
}

// AddEndorsement increments the endorsement count
func (s *UserSkill) AddEndorsement() {
	s.Endorsements++
	s.UpdatedAt = time.Now()
}

// UpdateNotes updates the skill notes
func (s *UserSkill) UpdateNotes(notes string) {
	s.Notes = notes
	s.UpdatedAt = time.Now()
}

// IsValid performs validation on the skill
func (s *UserSkill) IsValid() error {
	if s.Username == "" {
		return errors.ErrRequiredField
	}

	if s.SkillName == "" {
		return errors.ErrRequiredField
	}

	if !validProficiencyLevels[s.ProficiencyLevel] {
		return apperrors.ErrInvalidProficiencyLevel
	}

	if s.YearsOfExperience < 0 {
		return apperrors.ErrInvalidYearsOfExperience
	}

	return nil
}
