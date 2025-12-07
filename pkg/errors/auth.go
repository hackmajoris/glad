package errors

import "errors"

// Authentication errors - reusable across all apps in the monorepo
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)
