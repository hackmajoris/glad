package errors

import (
	"errors"
	"fmt"
)

// Is Core error utilities for the entire monorepo
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// ErrorType categorizes errors for better handling
type ErrorType int

// AppError provides structured error information
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (a *AppError) Error() string {
	if a.Err != nil {
		return fmt.Sprintf("%s: %v", a.Message, a.Err)
	}
	return a.Message
}

func (a *AppError) Unwrap() error {
	return a.Err
}
