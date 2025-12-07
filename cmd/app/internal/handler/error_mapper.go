package handler

import (
	"net/http"

	apperrors "github.com/hackmajoris/glad/cmd/app/internal/errors"
	pkgerrors "github.com/hackmajoris/glad/pkg/errors"
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

	// Validation errors
	case pkgerrors.Is(err, pkgerrors.ErrRequiredField):
		return http.StatusBadRequest, "Required field missing"
	case pkgerrors.Is(err, apperrors.ErrInvalidUsername):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidName):
		return http.StatusBadRequest, err.Error()
	case pkgerrors.Is(err, apperrors.ErrInvalidPassword):
		return http.StatusBadRequest, err.Error()

	// Default: Internal server error
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
