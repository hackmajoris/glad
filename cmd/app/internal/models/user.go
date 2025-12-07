package models

import (
	"fmt"
	"time"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	"github.com/hackmajoris/glad/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system (domain model)
// This entity uses single table design with the following key structure:
//   - PK: USER#<username>
//   - SK: PROFILE
type User struct {
	// Business attributes
	Username     string    `json:"username" dynamodbav:"Username"`
	Name         string    `json:"name" dynamodbav:"Name"`
	PasswordHash string    `json:"-" dynamodbav:"PasswordHash"`
	Email        string    `json:"email,omitempty" dynamodbav:"Email,omitempty"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time `json:"updated_at" dynamodbav:"UpdatedAt"`

	// DynamoDB system attributes for single table design
	PK         string `json:"-" dynamodbav:"PK"`
	SK         string `json:"-" dynamodbav:"SK"`
	EntityType string `json:"entity_type" dynamodbav:"EntityType"`
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
	user := &User{
		Username:     username,
		Name:         name,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
		EntityType:   "User",
	}

	// Set DynamoDB keys
	user.SetKeys()

	return user, nil
}

// SetKeys configures the PK and SK for DynamoDB single table design
func (u *User) SetKeys() {
	// Base table keys: User profile uses a fixed SK of "PROFILE"
	u.PK = fmt.Sprintf("USER#%s", u.Username)
	u.SK = "PROFILE"
	u.EntityType = "User"
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
