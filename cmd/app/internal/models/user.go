package models

import (
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system (domain model)
type User struct {
	Username     string    `json:"username" dynamodbav:"username"`
	Name         string    `json:"name" dynamodbav:"name"`
	PasswordHash string    `json:"-" dynamodbav:"password"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// NewUser creates a new User with the given credentials
func NewUser(username, name, password string) (*User, error) {
	if username == "" || password == "" || name == "" {
		return nil, errors.ErrRequiredField
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Username:     username,
		Name:         name,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateName updates the user's name
func (u *User) UpdateName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return apperrors.ErrInvalidName
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(password string) error {
	if len(password) < 6 {
		return apperrors.ErrInvalidPassword
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// ValidatePassword checks if the provided password matches the user's password
func (u *User) ValidatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

// GetUsername returns the username (implements auth.User interface)
func (u *User) GetUsername() string {
	return u.Username
}
