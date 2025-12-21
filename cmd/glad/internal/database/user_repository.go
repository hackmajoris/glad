package database

import "github.com/hackmajoris/glad-stack/cmd/glad/internal/models"

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUser(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	UserExists(username string) (bool, error)
	ListUsers() ([]*models.User, error)
}
