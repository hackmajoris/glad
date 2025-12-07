package database

import (
	"sync"
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"
)

// UserMockRepository implements UserRepository for testing
type UserMockRepository struct {
	users map[string]*models.User
	mutex sync.RWMutex
}

type UserSkillsMockRepository struct {
	users map[string]*models.UserSkill
	mutex sync.RWMutex
}

// NewUserMockRepository creates a new mock repository
func NewUserMockRepository() *UserMockRepository {
	log := logger.WithComponent("database")
	log.Info("Initializing Mock repository for local development")

	repo := &UserMockRepository{
		users: make(map[string]*models.User),
	}

	log.Info("Mock repository initialized successfully")
	return repo
}

// NewUserSkillsMockRepository NewUserMockRepository creates a new mock repository
func NewUserSkillsMockRepository() *UserSkillsMockRepository {
	log := logger.WithComponent("database")
	log.Info("Initializing Mock repository for local development")

	repo := &UserSkillsMockRepository{
		users: make(map[string]*models.UserSkill),
	}

	log.Info("Mock repository initialized successfully")
	return repo
}

// CreateUser creates a user in memory
func (m *UserMockRepository) CreateUser(user *models.User) error {
	log := logger.WithComponent("database").With("operation", "CreateUser", "username", user.Username, "repository", "mock")
	start := time.Now()

	log.Debug("Starting user creation in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.users[user.Username]; exists {
		log.Debug("User already exists", "duration", time.Since(start))
		return apperrors.ErrUserExists
	}

	m.users[user.Username] = user
	log.Info("User created successfully in mock repository", "total_users", len(m.users), "duration", time.Since(start))
	return nil
}

// GetUser retrieves a user from memory
func (m *UserMockRepository) GetUser(username string) (*models.User, error) {
	log := logger.WithComponent("database").With("operation", "GetUser", "username", username, "repository", "mock")
	start := time.Now()

	log.Debug("Starting user retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	user, exists := m.users[username]
	if !exists {
		log.Debug("User not found in mock repository", "duration", time.Since(start))
		return nil, apperrors.ErrUserNotFound
	}

	log.Debug("User retrieved successfully from mock repository", "duration", time.Since(start))
	return user, nil
}

// UpdateUser updates a user in memory
func (m *UserMockRepository) UpdateUser(user *models.User) error {
	log := logger.WithComponent("database").With("operation", "UpdateUser", "username", user.Username, "repository", "mock")
	start := time.Now()

	log.Debug("Starting user update in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.users[user.Username]; !exists {
		log.Debug("User not found for update", "duration", time.Since(start))
		return apperrors.ErrUserNotFound
	}

	m.users[user.Username] = user
	log.Info("User updated successfully in mock repository", "duration", time.Since(start))
	return nil
}

// UserExists checks if a user exists in memory
func (m *UserMockRepository) UserExists(username string) (bool, error) {
	log := logger.WithComponent("database").With("operation", "UserExists", "username", username, "repository", "mock")
	start := time.Now()

	log.Debug("Checking if user exists in mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.users[username]
	log.Debug("User existence check completed", "exists", exists, "duration", time.Since(start))
	return exists, nil
}

// ListUsers retrieves all users from memory
func (m *UserMockRepository) ListUsers() ([]*models.User, error) {
	log := logger.WithComponent("database").With("operation", "ListUsers", "repository", "mock")
	start := time.Now()

	log.Debug("Starting users list retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var users []*models.User
	for _, user := range m.users {
		users = append(users, user)
	}

	log.Info("Users retrieved successfully from mock repository", "count", len(users), "duration", time.Since(start))
	return users, nil
}

// CreateSkill creates a user skill in memory
func (m *UserSkillsMockRepository) CreateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "CreateSkill", "username", skill.Username, "skill", skill.SkillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill creation in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := skill.Username + "#" + skill.SkillName
	if _, exists := m.users[key]; exists {
		log.Debug("Skill already exists", "duration", time.Since(start))
		return apperrors.ErrSkillAlreadyExists
	}

	m.users[key] = skill
	log.Info("Skill created successfully in mock repository", "total_skills", len(m.users), "duration", time.Since(start))
	return nil
}

// GetSkill retrieves a user skill from memory
func (m *UserSkillsMockRepository) GetSkill(username, skillName string) (*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "GetSkill", "username", username, "skill", skillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill retrieval from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := username + "#" + skillName
	skill, exists := m.users[key]
	if !exists {
		log.Debug("Skill not found in mock repository", "duration", time.Since(start))
		return nil, apperrors.ErrSkillNotFound
	}

	log.Debug("Skill retrieved successfully from mock repository", "duration", time.Since(start))
	return skill, nil
}

// UpdateSkill updates a user skill in memory
func (m *UserSkillsMockRepository) UpdateSkill(skill *models.UserSkill) error {
	log := logger.WithComponent("database").With("operation", "UpdateSkill", "username", skill.Username, "skill", skill.SkillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill update in mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := skill.Username + "#" + skill.SkillName
	if _, exists := m.users[key]; !exists {
		log.Debug("Skill not found for update", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	m.users[key] = skill
	log.Info("Skill updated successfully in mock repository", "duration", time.Since(start))
	return nil
}

// DeleteSkill deletes a user skill from memory
func (m *UserSkillsMockRepository) DeleteSkill(username, skillName string) error {
	log := logger.WithComponent("database").With("operation", "DeleteSkill", "username", username, "skill", skillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skill deletion from mock repository")

	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := username + "#" + skillName
	if _, exists := m.users[key]; !exists {
		log.Debug("Skill not found for deletion", "duration", time.Since(start))
		return apperrors.ErrSkillNotFound
	}

	delete(m.users, key)
	log.Info("Skill deleted successfully from mock repository", "duration", time.Since(start))
	return nil
}

// ListSkillsForUser retrieves all skills for a specific user from memory
func (m *UserSkillsMockRepository) ListSkillsForUser(username string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListSkillsForUser", "username", username, "repository", "mock")
	start := time.Now()

	log.Debug("Starting skills list retrieval for user from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for key, skill := range m.users {
		if skill.Username == username {
			skills = append(skills, skill)
		}
		_ = key
	}

	log.Info("Skills retrieved successfully for user from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkill retrieves all users with a specific skill from memory
func (m *UserSkillsMockRepository) ListUsersBySkill(skillName string) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkill", "skill", skillName, "repository", "mock")
	start := time.Now()

	log.Debug("Starting users list retrieval by skill from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for key, skill := range m.users {
		if skill.SkillName == skillName {
			skills = append(skills, skill)
		}
		_ = key
	}

	log.Info("Users retrieved successfully by skill from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}

// ListUsersBySkillAndLevel retrieves all users with a specific skill and proficiency level from memory
func (m *UserSkillsMockRepository) ListUsersBySkillAndLevel(skillName string, proficiencyLevel models.ProficiencyLevel) ([]*models.UserSkill, error) {
	log := logger.WithComponent("database").With("operation", "ListUsersBySkillAndLevel", "skill", skillName, "level", proficiencyLevel, "repository", "mock")
	start := time.Now()

	log.Debug("Starting users list retrieval by skill and level from mock repository")

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var skills []*models.UserSkill
	for key, skill := range m.users {
		if skill.SkillName == skillName && skill.ProficiencyLevel == proficiencyLevel {
			skills = append(skills, skill)
		}
		_ = key
	}

	log.Info("Users retrieved successfully by skill and level from mock repository", "count", len(skills), "duration", time.Since(start))
	return skills, nil
}
