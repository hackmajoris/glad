package validation

import (
	apperrors "github.com/hackmajoris/glad-stack/cmd/glad/internal/errors"
)

// Validator provides validation functionality
type Validator struct{}

// New creates a new Validator
func New() *Validator {
	return &Validator{}
}

// ValidateOptionalName validates an optional name (for updates)
func (v *Validator) ValidateOptionalName(name *string) error {
	if name == nil {
		return nil
	}
	if len(*name) < 2 || len(*name) > 100 {
		return apperrors.ErrInvalidName
	}
	return nil
}
