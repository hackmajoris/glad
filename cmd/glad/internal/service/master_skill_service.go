package service

import (
	"time"

	"github.com/hackmajoris/glad/cmd/glad/internal/database"
	"github.com/hackmajoris/glad/cmd/glad/internal/dto"
	"github.com/hackmajoris/glad/cmd/glad/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"
)

// MasterSkillService handles master skill business logic
type MasterSkillService struct {
	repo database.MasterSkillRepository
}

// NewMasterSkillService creates a new MasterSkillService
func NewMasterSkillService(repo database.MasterSkillRepository) *MasterSkillService {
	return &MasterSkillService{
		repo: repo,
	}
}

// CreateMasterSkill creates a new master skill
func (s *MasterSkillService) CreateMasterSkill(skillID, skillName, description, category string, tags []string) (*models.Skill, error) {
	log := logger.WithComponent("service").With("operation", "CreateMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Info("Processing create master skill request")

	// Create new master skill
	skill, err := models.NewSkill(skillID, skillName, description, category, tags)
	if err != nil {
		log.Error("Failed to create skill model", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Save to database
	if err := s.repo.CreateMasterSkill(skill); err != nil {
		log.Error("Failed to save master skill to database", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("Master skill created successfully", "duration", time.Since(start))
	return skill, nil
}

// GetMasterSkill retrieves a master skill by ID
func (s *MasterSkillService) GetMasterSkill(skillID string) (*models.Skill, error) {
	log := logger.WithComponent("service").With("operation", "GetMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Debug("Retrieving master skill")

	skill, err := s.repo.GetMasterSkill(skillID)
	if err != nil {
		log.Error("Failed to get master skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Debug("Master skill retrieved successfully", "duration", time.Since(start))
	return skill, nil
}

// UpdateMasterSkill updates an existing master skill
func (s *MasterSkillService) UpdateMasterSkill(skillID, skillName, description, category string, tags []string) (*models.Skill, error) {
	log := logger.WithComponent("service").With("operation", "UpdateMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Info("Processing update master skill request")

	// Get existing skill
	skill, err := s.repo.GetMasterSkill(skillID)
	if err != nil {
		log.Error("Failed to get master skill", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Update fields if provided
	if skillName != "" || description != "" || category != "" {
		skill.UpdateMetadata(skillName, description, category)
	}

	if tags != nil {
		skill.UpdateTags(tags)
	}

	// Save updated skill
	if err := s.repo.UpdateMasterSkill(skill); err != nil {
		log.Error("Failed to update master skill in database", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	log.Info("Master skill updated successfully", "duration", time.Since(start))
	return skill, nil
}

// DeleteMasterSkill deletes a master skill
func (s *MasterSkillService) DeleteMasterSkill(skillID string) error {
	log := logger.WithComponent("service").With("operation", "DeleteMasterSkill", "skill_id", skillID)
	start := time.Now()

	log.Info("Processing delete master skill request")

	if err := s.repo.DeleteMasterSkill(skillID); err != nil {
		log.Error("Failed to delete master skill", "error", err.Error(), "duration", time.Since(start))
		return err
	}

	log.Info("Master skill deleted successfully", "duration", time.Since(start))
	return nil
}

// ListMasterSkills retrieves all master skills
func (s *MasterSkillService) ListMasterSkills() ([]dto.MasterSkillResponse, error) {
	log := logger.WithComponent("service").With("operation", "ListMasterSkills")
	start := time.Now()

	log.Info("Retrieving all master skills")

	skills, err := s.repo.ListMasterSkills()
	if err != nil {
		log.Error("Failed to retrieve master skills", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	// Convert to response DTOs
	result := make([]dto.MasterSkillResponse, len(skills))
	for i, skill := range skills {
		result[i] = dto.MasterSkillResponse{
			SkillID:     skill.SkillID,
			SkillName:   skill.SkillName,
			Description: skill.Description,
			Category:    skill.Category,
			Tags:        skill.Tags,
			CreatedAt:   skill.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   skill.UpdatedAt.Format(time.RFC3339),
		}
	}

	log.Info("Master skills retrieved successfully", "count", len(result), "duration", time.Since(start))
	return result, nil
}
