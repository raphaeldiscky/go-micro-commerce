package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
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
		err := validate.Struct(test1)
		assert.NoError(t, err)

		// Should fail with zero
		test2 := TestGT{Price: decimal.Zero}
		err = validate.Struct(test2)
		assert.Error(t, err)

		// Should fail with negative value
		test3 := TestGT{Price: decimal.NewFromFloat(-5.00)}
		err = validate.Struct(test3)
		assert.Error(t, err)
	})

	t.Run("decimal_gte validator", func(t *testing.T) {
		t.Parallel()
		// Should pass with positive value
		test1 := TestGTE{Price: decimal.NewFromFloat(10.50)}
		err := validate.Struct(test1)
		assert.NoError(t, err)

		// Should pass with zero
		test2 := TestGTE{Price: decimal.Zero}
		err = validate.Struct(test2)
		assert.NoError(t, err)

		// Should fail with negative value
		test3 := TestGTE{Price: decimal.NewFromFloat(-5.00)}
		err = validate.Struct(test3)
		assert.Error(t, err)
	})
}
