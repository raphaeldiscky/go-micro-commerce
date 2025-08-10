package validationutils

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// DecimalGT validates that the field is greater than the given parameter.
func DecimalGT(fl validator.FieldLevel) bool {
	data, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	value, err := decimal.NewFromString(data)
	if err != nil {
		return false
	}

	baseValue, err := decimal.NewFromString(fl.Param())
	if err != nil {
		return false
	}

	return value.GreaterThan(baseValue)
}

// DecimalLT validates that the field is less than the given parameter.
func DecimalLT(fl validator.FieldLevel) bool {
	data, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	value, err := decimal.NewFromString(data)
	if err != nil {
		return false
	}

	baseValue, err := decimal.NewFromString(fl.Param())
	if err != nil {
		return false
	}

	return value.LessThan(baseValue)
}

// DecimalGTE validates that the field is greater than or equal to the given parameter.
func DecimalGTE(fl validator.FieldLevel) bool {
	data, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	value, err := decimal.NewFromString(data)
	if err != nil {
		return false
	}

	baseValue, err := decimal.NewFromString(fl.Param())
	if err != nil {
		return false
	}

	return value.GreaterThanOrEqual(baseValue)
}

// DecimalLTE validates that the field is less than or equal to the given parameter.
func DecimalLTE(fl validator.FieldLevel) bool {
	data, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	value, err := decimal.NewFromString(data)
	if err != nil {
		return false
	}

	baseValue, err := decimal.NewFromString(fl.Param())
	if err != nil {
		return false
	}

	return value.LessThanOrEqual(baseValue)
}
