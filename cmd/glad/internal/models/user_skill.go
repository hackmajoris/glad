package models

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/glad/internal/errors"
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
// This entity uses single table design with multi-attribute composite keys:
//   - entity_id: USERSKILL#<username>#<skill_id>
//   - skill_id: Immutable skill reference (e.g., "python")
//   - SkillName: Denormalized display name for GSI queries
//   - Category: Denormalized from master Skill
//
// GSI SkillsByLevel uses: SkillName + ProficiencyLevel + YearsOfExperience + Username
// GSI ByUser uses: Username + EntityType
type UserSkill struct {
	// Business attributes - used directly in GSI composite keys
	Username          string           `json:"username" dynamodbav:"Username"`
	SkillID           string           `json:"skill_id" dynamodbav:"skill_id"`    // Immutable reference
	SkillName         string           `json:"skill_name" dynamodbav:"SkillName"` // Denormalized for GSI
	Category          string           `json:"category" dynamodbav:"Category"`    // Denormalized from Skill
	ProficiencyLevel  ProficiencyLevel `json:"proficiency_level" dynamodbav:"ProficiencyLevel"`
	YearsOfExperience int              `json:"years_of_experience" dynamodbav:"YearsOfExperience"`
	Endorsements      int              `json:"endorsements" dynamodbav:"Endorsements"`
	LastUsedDate      string           `json:"last_used_date" dynamodbav:"LastUsedDate"` // ISO 8601 format
	Notes             string           `json:"notes,omitempty" dynamodbav:"Notes,omitempty"`
	CreatedAt         time.Time        `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt         time.Time        `json:"updated_at" dynamodbav:"UpdatedAt"`

	// DynamoDB attributes
	EntityID           string `json:"-" dynamodbav:"entity_id"`
	EntityType         string `json:"entity_type" dynamodbav:"EntityType"`
	SkillCompositeSort string `json:"-" dynamodbav:"SkillCompositeSort"`
}

// NewUserSkill creates a new UserSkill with proper validation
// skillID: Immutable skill identifier (e.g., "python")
// skillName: Display name (e.g., "Python") - denormalized from master Skill
// category: Skill category (e.g., "Programming") - denormalized from master Skill
func NewUserSkill(username, skillID, skillName, category string, proficiencyLevel ProficiencyLevel, yearsOfExperience int) (*UserSkill, error) {
	if username == "" {
		return nil, errors.ErrRequiredField
	}

	if skillID == "" || skillName == "" {
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
		SkillID:           skillID,
		SkillName:         skillName,
		Category:          category,
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

func (s *UserSkill) SetKeys() {
	// Base table key: Unique identifier
	s.EntityID = BuildUserSkillEntityID(s.Username, s.SkillID)
	s.EntityType = "UserSkill"
}

// UpdateProficiency updates the skill proficiency level
func (s *UserSkill) UpdateProficiency(level ProficiencyLevel) error {
	if !validProficiencyLevels[level] {
		return apperrors.ErrInvalidProficiencyLevel
	}

	s.ProficiencyLevel = level
	s.UpdatedAt = time.Now()

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
