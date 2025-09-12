// Package validationutils is a utility package for validation-related functions
package validationutils

import (
	"reflect"

	"github.com/shopspring/decimal"
)

// DecimalType converts a decimal.Decimal to its string representation.
func DecimalType(field reflect.Value) any {
	if valuer, ok := field.Interface().(decimal.Decimal); ok {
		return valuer.String()
	}

	return nil
}
