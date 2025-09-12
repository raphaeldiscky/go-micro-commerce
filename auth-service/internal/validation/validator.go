// Package validation provides custom validation logic for Echo framework.
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator instance.
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator initializes a new CustomValidator with the go-playground validator.
func NewValidator() echo.Validator {
	validate := validator.New()

	// Register custom validation rules here if needed

	return &CustomValidator{validator: validate}
}

// Validate implements the echo.Validator interface.
func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
