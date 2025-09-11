package validation_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestDecimalValidators(t *testing.T) {
	t.Parallel()
	// Create validator instance
	validate := validator.New()

	// Register decimal_gt validator
	err := validate.RegisterValidation("decimal_gt", func(fl validator.FieldLevel) bool {
		if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
			return dec.GreaterThan(decimal.Zero)
		}

		return false
	})
	require.NoError(t, err)

	// Register decimal_gte validator
	err = validate.RegisterValidation("decimal_gte", func(fl validator.FieldLevel) bool {
		if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
			return dec.GreaterThanOrEqual(decimal.Zero)
		}

		return false
	})
	require.NoError(t, err)

	// Test struct for decimal_gt
	type TestGT struct {
		Price decimal.Decimal `validate:"decimal_gt"`
	}

	// Test struct for decimal_gte
	type TestGTE struct {
		Price decimal.Decimal `validate:"decimal_gte"`
	}

	t.Run("decimal_gt validator", func(t *testing.T) {
		t.Parallel()
		// Should pass with positive value
		test1 := TestGT{Price: decimal.NewFromFloat(10.50)}
		validationErr := validate.Struct(test1)
		require.NoError(t, validationErr)

		// Should fail with zero
		test2 := TestGT{Price: decimal.Zero}
		validationErr = validate.Struct(test2)
		require.Error(t, validationErr)

		// Should fail with negative value
		test3 := TestGT{Price: decimal.NewFromFloat(-5.00)}
		validationErr = validate.Struct(test3)
		require.Error(t, validationErr)
	})

	t.Run("decimal_gte validator", func(t *testing.T) {
		t.Parallel()
		// Should pass with positive value
		test1 := TestGTE{Price: decimal.NewFromFloat(10.50)}
		validationErr := validate.Struct(test1)
		require.NoError(t, validationErr)

		// Should pass with zero
		test2 := TestGTE{Price: decimal.Zero}
		validationErr = validate.Struct(test2)
		require.NoError(t, validationErr)

		// Should fail with negative value
		test3 := TestGTE{Price: decimal.NewFromFloat(-5.00)}
		validationErr = validate.Struct(test3)
		require.Error(t, validationErr)
	})
}
