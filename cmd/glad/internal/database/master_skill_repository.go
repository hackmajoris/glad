package database

import "github.com/hackmajoris/glad/cmd/glad/internal/models"

// MasterSkillRepository defines operations for master skills
type MasterSkillRepository interface {
	CreateMasterSkill(skill *models.Skill) error
	GetMasterSkill(skillID string) (*models.Skill, error)
	UpdateMasterSkill(skill *models.Skill) error
	DeleteMasterSkill(skillID string) error
	ListMasterSkills() ([]*models.Skill, error)
}
