package database

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/cmd/app/internal/models"
	"github.com/hackmajoris/glad/pkg/logger"
)

// CreateUser creates a user in memory
func (m *MockRepository) CreateUser(user *models.User) error {
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
func (m *MockRepository) GetUser(username string) (*models.User, error) {
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
func (m *MockRepository) UpdateUser(user *models.User) error {
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
func (m *MockRepository) UserExists(username string) (bool, error) {
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
func (m *MockRepository) ListUsers() ([]*models.User, error) {
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
