package database

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"
)

// CreateSkill creates a user skill in memory
func (m *MockRepository) CreateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "CreateSkill", "username", skill.Username, "skill_id", skill.SkillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill creation in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := models.BuildUserSkillEntityID(skill.Username, skill.SkillID)
	if _, exists := m.skills[key]; exists {
		log.Debug("Skill already exists", "duration", time.Since(start))
		return apperrors.ErrSkillAlreadyExists
	}

	m.skills[key] = skill
	log.Info("Skill created successfully in mock repository", "total_skills", len(m.skills), "duration", time.Since(start))
	return nil
}

// GetSkill retrieves a user skill from memory
func (m *MockRepository) GetSkill(username, skillID string) (*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "GetSkill", "username", username, "skill_id", skillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := models.BuildUserSkillEntityID(username, skillID)
	skill, exists := m.skills[key]
	if !exists {
		log.Debug("Skill not found in mock repository", "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	log.Debug("Skill retrieved successfully from mock repository", "duration", time.Since(start))
	return skill, nil
}

// UpdateSkill updates a user skill in memory
func (m *MockRepository) UpdateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "UpdateSkill", "username", skill.Username, "skill_id", skill.SkillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill update in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := models.BuildUserSkillEntityID(skill.Username, skill.SkillID)
	if _, exists := m.skills[key]; !exists {
		log.Debug("Skill not found for update", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	m.skills[key] = skill
	log.Info("Skill updated successfully in mock repository", "duration", time.Since(start))
	return nil
}

// DeleteSkill deletes a user skill from memory
func (m *MockRepository) DeleteSkill(username, skillID string) error {
	log := logger.WithComponent("database").With("operation", "DeleteSkill", "username", username, "skill_id", skillID, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill deletion from mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := models.BuildUserSkillEntityID(username, skillID)
	if _, exists := m.skills[key]; !exists {
		log.Debug("Skill not found for deletion", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	delete(m.skills, key)
	log.Info("Skill deleted successfully from mock repository", "duration", time.Since(start))
	return nil
}

// ListSkillsForUser retrieves all skills for a specific user from memory
func (m *MockRepository) ListSkillsForUser(username string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListSkillsForUser", "username", username, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skills list retrieval for user from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for _, skill := range m.skills {
		if skill.Username == username {
			skills = append(skills, skill)
		}
	}

	log.Info("Skills retrieved successfully for user from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkill retrieves all users with a specific skill from memory
func (m *MockRepository) ListUsersBySkill(skillName string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkill", "skill", skillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting users list retrieval by skill from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for _, skill := range m.skills {
		if skill.SkillName == skillName {
			skills = append(skills, skill)
		}
	}

	log.Info("Users retrieved successfully by skill from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkillAndLevel retrieves all users with a specific skill and proficiency level from memory
func (m *MockRepository) ListUsersBySkillAndLevel(skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkillAndLevel", "skill", skillName, "level", proficiencyLevel, "repository", "mock")
	start := time.Now()

	log.Debug("Starting users list retrieval by skill and level from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for _, skill := range m.skills {
		if skill.SkillName == skillName && skill.ProficiencyLevel == proficiencyLevel {
			skills = append(skills, skill)
		}
	}

	log.Info("Users retrieved successfully by skill and level from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
