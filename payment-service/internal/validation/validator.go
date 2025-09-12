// Package validation provides custom validation rules for the HTTP server.
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

// CustomValidator wraps the go-playground validator.
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator initializes a validator with custom rules.
func NewValidator() echo.Validator {
	validate := validator.New()

	// Register custom validators
	err := validate.RegisterValidation("decimal_gt", decimalGT)
	if err != nil {
		panic("failed to register decimal_gt validator: " + err.Error())
	}

	err = validate.RegisterValidation("decimal_gte", decimalGTE)
	if err != nil {
		panic("failed to register decimal_gte validator: " + err.Error())
	}

	return &CustomValidator{validator: validate}
}

// Validate implements the echo.Validator interface.
func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

// decimalGT checks if decimal > 0.
func decimalGT(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return dec.GreaterThan(decimal.Zero)
	}

	return false
}

// decimalGTE checks if decimal >= 0.
func decimalGTE(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return dec.GreaterThanOrEqual(decimal.Zero)
	}

	return false
}
