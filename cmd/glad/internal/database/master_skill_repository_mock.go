package database

import (
	"time"

	apperrors "github.com/hackmajoris/glad-stack/cmd/glad/internal/errors"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/pkg/logger"
)

// CreateMasterSkill creates a master skill in memory
func (m *MockRepository) CreateMasterSkill(skill *models.Skill) error {
	log := logger.WithComponent("database").With("operation", "CreateMasterSkill", "skill_id", skill.SkillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting master skill creation in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.masterSkills[skill.SkillID]; exists {
		log.Debug("Master skill already exists", "duration", time.Since(start))
		return apperrors.ErrSkillAlreadyExists
	}

	m.masterSkills[skill.SkillID] = skill
	log.Info("Master skill created successfully in mock repository", "total_master_skills", len(m.masterSkills), "duration", time.Since(start))
	return nil
}

// GetMasterSkill retrieves a master skill from memory
func (m *MockRepository) GetMasterSkill(skillID string) (*models.Skill, error) {
	log := logger.WithComponent("database").With("operation", "GetMasterSkill", "skill_id", skillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting master skill retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	skill, exists := m.masterSkills[skillID]
	if !exists {
		log.Debug("Master skill not found in mock repository", "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	log.Debug("Master skill retrieved successfully from mock repository", "duration", time.Since(start))
	return skill, nil
}

// UpdateMasterSkill updates a master skill in memory
func (m *MockRepository) UpdateMasterSkill(skill *models.Skill) error {
	log := logger.WithComponent("database").With("operation", "UpdateMasterSkill", "skill_id", skill.SkillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting master skill update in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.masterSkills[skill.SkillID]; !exists {
		log.Debug("Master skill not found for update", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	m.masterSkills[skill.SkillID] = skill
	log.Info("Master skill updated successfully in mock repository", "duration", time.Since(start))
	return nil
}

// DeleteMasterSkill deletes a master skill from memory
func (m *MockRepository) DeleteMasterSkill(skillID string) error {
	log := logger.WithComponent("database").With("operation", "DeleteMasterSkill", "skill_id", skillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting master skill deletion from mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.masterSkills[skillID]; !exists {
		log.Debug("Master skill not found for deletion", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	delete(m.masterSkills, skillID)
	log.Info("Master skill deleted successfully from mock repository", "duration", time.Since(start))
	return nil
}

// ListMasterSkills retrieves all master skills from memory
func (m *MockRepository) ListMasterSkills() ([]*models.Skill, error) {
	log := logger.WithComponent("database").With("operation", "ListMasterSkills", "repository", "mock")
	start := time.Now()

	log.Debug("Starting master skills list retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.Skill
	for _, skill := range m.masterSkills {
		skills = append(skills, skill)
	}

	log.Info("Master skills retrieved successfully from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
