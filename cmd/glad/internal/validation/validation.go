package validation

import (
	apperrors "github.com/hackmajoris/glad/cmd/glad/internal/errors"
	pkgerrors "github.com/hackmajoris/glad/pkg/errors"
)

// Validator provides validation functionality
type Validator struct{}

// New creates a new Validator
func New() *Validator {
	return &Validator{}
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return pkgerrors.ErrRequiredField
	}
	if len(username) < 3 || len(username) > 50 {
		return apperrors.ErrInvalidUsername
	}
	return nil
}

// ValidateName validates a name
func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return pkgerrors.ErrRequiredField
	}
	if len(name) < 2 || len(name) > 100 {
		return apperrors.ErrInvalidName
	}
	return nil
}

// ValidatePassword validates a password
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return pkgerrors.ErrRequiredField
	}
	if len(password) < 6 {
		return apperrors.ErrInvalidPassword
	}
	return nil
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

// ValidateOptionalPassword validates an optional password (for updates)
func (v *Validator) ValidateOptionalPassword(password *string) error {
	if password == nil {
		return nil
	}
	if len(*password) < 6 {
		return apperrors.ErrInvalidPassword
	}
	return nil
}

// ValidateRegisterInput validates registration input
func (v *Validator) ValidateRegisterInput(username, name, password string) error {
	if err := v.ValidateUsername(username); err != nil {
		return err
	}
	if err := v.ValidateName(name); err != nil {
		return err
	}
	if err := v.ValidatePassword(password); err != nil {
		return err
	}
	return nil
}

// ValidateLoginInput validates login input
func (v *Validator) ValidateLoginInput(username, password string) error {
	if username == "" {
		return pkgerrors.ErrRequiredField
	}
	if password == "" {
		return pkgerrors.ErrRequiredField
	}
	return nil
}
