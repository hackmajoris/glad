package database

import "github.com/hackmajoris/glad/cmd/app/internal/models"

// SkillRepository defines operations for user skills
type SkillRepository interface {
	CreateSkill(skill *models.UserSkill) error
	GetSkill(username, skillID string) (*models.UserSkill, error)
	UpdateSkill(skill *models.UserSkill) error
	DeleteSkill(username, skillID string) error
	ListSkillsForUser(username string) ([]*models.UserSkill, error)
	ListUsersBySkill(skillName string) ([]*models.UserSkill, error)
	ListUsersBySkillAndLevel(skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error)
}
