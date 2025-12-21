package service

import (
	"time"

	"github.com/hackmajoris/glad/cmd/glad/internal/database"
	"github.com/hackmajoris/glad/cmd/glad/internal/dto"
	apperrors "github.com/hackmajoris/glad/cmd/glad/internal/errors"
	"github.com/hackmajoris/glad/cmd/glad/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"
)

// Re-export domain errors for convenience in handler layer
var (
	ErrSkillNotFound            = apperrors.ErrSkillNotFound
	ErrSkillAlreadyExists       = apperrors.ErrSkillAlreadyExists
	ErrInvalidProficiencyLevel  = apperrors.ErrInvalidProficiencyLevel
	ErrInvalidYearsOfExperience = apperrors.ErrInvalidYearsOfExperience
	ErrInvalidSkillName         = apperrors.ErrInvalidSkillName
)

// SkillService handles skill business logic
type SkillService struct {
	repo            database.SkillRepository
	masterSkillRepo database.MasterSkillRepository
	userRepo        database.UserRepository
}

// NewSkillService creates a new SkillService
func NewSkillService(repo database.SkillRepository, masterSkillRepo database.MasterSkillRepository, userRepo database.UserRepository) *SkillService {
	return &SkillService{
		repo:            repo,
		masterSkillRepo: masterSkillRepo,
		userRepo:        userRepo,
	}
}

// AddSkill adds a new skill to a user
// The skillName parameter is used as the skillID to look up the master skill
func (s *SkillService) AddSkill(username, skillName string, proficiencyLevel models.ProficiencyLevel, yearsOfExperience int, notes string) (*models.UserSkill, error) {
	log := logger.WithComponent("service").With("operation", "AddSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Info("Processing add skill request")

	// Look up master skill to get skillID, skillName, and category
	masterSkill, err := s.masterSkillRepo.GetMasterSkill(skillName)
	if err != nil {
		log.Error("Master skill not found", "error", err.Error(), "skill_id", skillName, "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	log.Debug("Master skill found", "skill_id", masterSkill.SkillID, "skill_name", masterSkill.SkillName, "category", masterSkill.Category)

	// Create new user skill with data from master skill
	skill, err := models.NewUserSkill(username, masterSkill.SkillID, masterSkill.SkillName, masterSkill.Category, proficiencyLevel, yearsOfExperience)
	if err != nil {
		log.Error("Failed to create skill model", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	if notes != "" {
		skill.UpdateNotes(notes)
	}

	// Save skill to database
	if err := s.repo.CreateSkill(skill); err != nil {
		log.Error("Failed to save skill to database", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("Skill added successfully", "duration", time.Since(start))
	return skill, nil
}

// GetSkill retrieves a specific skill for a user
func (s *SkillService) GetSkill(username, skillName string) (*models.UserSkill, error) {
	log := logger.WithComponent("service").With("operation", "GetSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Debug("Retrieving skill")

	skill, err := s.repo.GetSkill(username, skillName)
	if err != nil {
		log.Error("Failed to get skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Debug("Skill retrieved successfully", "duration", time.Since(start))
	return skill, nil
}

// UpdateSkill updates an existing skill
func (s *SkillService) UpdateSkill(username, skillName string, proficiencyLevel *models.ProficiencyLevel, yearsOfExperience *int, notes *string) (*models.UserSkill, error) {
	log := logger.WithComponent("service").With("operation", "UpdateSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Info("Processing update skill request")

	// Get existing skill
	skill, err := s.repo.GetSkill(username, skillName)
	if err != nil {
		log.Error("Failed to get skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Update fields if provided
	if proficiencyLevel != nil {
		if err := skill.UpdateProficiency(*proficiencyLevel); err != nil {
			log.Error("Failed to update proficiency level", "error", err.Error(), "duration", time.Since(start))
			return nil, err
		}
	}

	if yearsOfExperience != nil {
		if err := skill.UpdateYearsOfExperience(*yearsOfExperience); err != nil {
			log.Error("Failed to update years of experience", "error", err.Error(), "duration", time.Since(start))
			return nil, err
		}
	}

	if notes != nil {
		skill.UpdateNotes(*notes)
	}

	// Save updated skill
	if err := s.repo.UpdateSkill(skill); err != nil {
		log.Error("Failed to update skill in database", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("Skill updated successfully", "duration", time.Since(start))
	return skill, nil
}

// DeleteSkill removes a skill from a user
func (s *SkillService) DeleteSkill(username, skillName string) error {
	log := logger.WithComponent("service").With("operation", "DeleteSkill", "username", username, "skill", skillName)
	start := time.Now()

	log.Info("Processing delete skill request")

	if err := s.repo.DeleteSkill(username, skillName); err != nil {
		log.Error("Failed to delete skill", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("Skill deleted successfully", "duration", time.Since(start))
	return nil
}

// ListSkillsForUser retrieves all skills for a user
func (s *SkillService) ListSkillsForUser(username string) ([]dto.SkillResponse, error) {
	log := logger.WithComponent("service").With("operation", "ListSkillsForUser", "username", username)
	start := time.Now()

	log.Info("Retrieving skills for user")

	// Check if user exists
	if _, err := s.userRepo.GetUser(username); err != nil {
		log.Error("User not found", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	skills, err := s.repo.ListSkillsForUser(username)
	if err != nil {
		log.Error("Failed to retrieve skills", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Convert to response DTOs
	result := make([]dto.SkillResponse, len(skills))
	for i, skill := range skills {
		result[i] = dto.SkillResponse{
			SkillName:         skill.SkillName,
			ProficiencyLevel:  string(skill.ProficiencyLevel),
			YearsOfExperience: skill.YearsOfExperience,
			Endorsements:      skill.Endorsements,
			LastUsedDate:      skill.LastUsedDate,
			Notes:             skill.Notes,
			CreatedAt:         skill.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         skill.UpdatedAt.Format(time.RFC3339),
		}
	}

	log.Info("Skills retrieved successfully", "count", len(result), "duration", time.Since(start))
	return result, nil
}

// ListUsersBySkill retrieves all users who have a specific skill in a category
func (s *SkillService) ListUsersBySkill(category, skillName string) ([]dto.UserSkillResponse, error) {
	log := logger.WithComponent("service").With("operation", "ListUsersBySkill", "category", category, "skill", skillName)
	start := time.Now()

	log.Info("Retrieving users by skill")

	skills, err := s.repo.ListUsersBySkill(category, skillName)
	if err != nil {
		log.Error("Failed to retrieve users by skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Convert to response DTOs
	result := make([]dto.UserSkillResponse, len(skills))
	for i, skill := range skills {
		result[i] = dto.UserSkillResponse{
			Username:          skill.Username,
			SkillName:         skill.SkillName,
			ProficiencyLevel:  string(skill.ProficiencyLevel),
			YearsOfExperience: skill.YearsOfExperience,
			Endorsements:      skill.Endorsements,
			LastUsedDate:      skill.LastUsedDate,
		}
	}

	log.Info("Users with skill retrieved successfully", "category", category, "skill", skillName, "count", len(result), "duration", time.Since(start))
	return result, nil
}

// ListUsersBySkillAndLevel retrieves users with a skill at a specific proficiency level in a category
func (s *SkillService) ListUsersBySkillAndLevel(category, skillName string, proficiencyLevel models.ProficiencyLevel) ([]dto.UserSkillResponse, error) {
	log := logger.WithComponent("service").With("operation", "ListUsersBySkillAndLevel", "category", category, "skill", skillName, "level", proficiencyLevel)
	start := time.Now()

	log.Info("Retrieving users by skill and level")

	skills, err := s.repo.ListUsersBySkillAndLevel(category, skillName, proficiencyLevel)
	if err != nil {
		log.Error("Failed to retrieve users by skill and level", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Convert to response DTOs
	result := make([]dto.UserSkillResponse, len(skills))
	for i, skill := range skills {
		result[i] = dto.UserSkillResponse{
			Username:          skill.Username,
			SkillName:         skill.SkillName,
			ProficiencyLevel:  string(skill.ProficiencyLevel),
			YearsOfExperience: skill.YearsOfExperience,
			Endorsements:      skill.Endorsements,
			LastUsedDate:      skill.LastUsedDate,
		}
	}

	log.Info("Users with skill and level retrieved successfully", "category", category, "skill", skillName, "level", proficiencyLevel, "count", len(result), "duration", time.Since(start))
	return result, nil
}
