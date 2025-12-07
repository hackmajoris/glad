package errors

import "errors"

// User-related domain errors
var (
	// ErrUserExists User existence errors
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidUsername Validation errors
	ErrInvalidUsername = errors.New("username must be between 3 and 50 characters")
	ErrInvalidName     = errors.New("name must be between 2 and 100 characters")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")

	// ErrInvalidCredentials Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrSkillNotFound Skill-related errors
	ErrSkillNotFound            = errors.New("skill not found")
	ErrSkillAlreadyExists       = errors.New("skill already exists for this user")
	ErrInvalidProficiencyLevel  = errors.New("proficiency level must be Beginner, Intermediate, Advanced, or Expert")
	ErrInvalidYearsOfExperience = errors.New("years of experience must be non-negative")
	ErrInvalidSkillName         = errors.New("skill name must be between 1 and 100 characters")
)
