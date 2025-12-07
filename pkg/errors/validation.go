package errors

import (
	"errors"
	"fmt"
)

// Generic validation errors - reusable across all apps
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrRequiredField = errors.New("required field missing")
	ErrInvalidLength = errors.New("invalid field length")
)

// FieldValidationError provides detailed validation error information for specific fields
type FieldValidationError struct {
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (v *FieldValidationError) Error() string {
	if v.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", v.Field, v.Message)
	}
	return fmt.Sprintf("validation failed: %s", v.Message)
}
