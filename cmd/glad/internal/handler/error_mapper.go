package handler

import (
	"net/http"

	apperrors "github.com/hackmajoris/glad-stack/cmd/glad/internal/errors"
	pkgerrors "github.com/hackmajoris/glad-stack/pkg/errors"
)

// ErrorMapper maps service errors to HTTP status codes and messages
type ErrorMapper struct{}

// NewErrorMapper creates a new ErrorMapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapToHTTP converts a service error to HTTP status code and message
func (em *ErrorMapper) MapToHTTP(err error) (int, string) {
	switch {
	// User existence errors
	case pkgerrors.Is(err, apperrors.ErrUserNotFound):
		return http.StatusNotFound, "User not found"
	case pkgerrors.Is(err, apperrors.ErrUserExists):
		return http.StatusConflict, "User already exists"

	// Authentication errors
	case pkgerrors.Is(err, apperrors.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Invalid credentials"

	// Skill errors
	case pkgerrors.Is(err, apperrors.ErrSkillNotFound):
		return http.StatusNotFound, "Skill not found"
	case pkgerrors.Is(err, apperrors.ErrSkillAlreadyExists):
		return http.StatusConflict, "Skill already exists for this user"

	// Master skill errors
	case pkgerrors.Is(err, apperrors.ErrMasterSkillNotFound):
		return http.StatusNotFound, "Master skill not found"
	case pkgerrors.Is(err, apperrors.ErrMasterSkillExists):
		return http.StatusConflict, "Master skill already exists"

	// Validation errors
	case pkgerrors.Is(err, pkgerrors.ErrRequiredField):
		return http.StatusBadRequest, "Required field missing"
	case pkgerrors.Is(err, apperrors.ErrInvalidUsername):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidName):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidPassword):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidProficiencyLevel):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidYearsOfExperience):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidSkillName):
		return http.StatusBadRequest, err.Error()

	// Default: Internal server error
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
