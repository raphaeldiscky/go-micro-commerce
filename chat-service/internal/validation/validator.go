// Package validation provides custom validation rules for the HTTP server.
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the go-playground validator.
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator initializes a validator with custom rules.
func NewValidator() echo.Validator {
	validate := validator.New()
	return &CustomValidator{validator: validate}
}

// Validate implements the echo.Validator interface.
func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
