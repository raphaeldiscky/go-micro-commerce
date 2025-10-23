// Package scalar provides custom GraphQL scalar type implementations.
package scalar

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/shopspring/decimal"
)

// MarshalDecimal serializes a decimal.Decimal to GraphQL as a quoted string.
// This maintains precision by avoiding floating-point representation.
func MarshalDecimal(d decimal.Decimal) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		// Write as quoted string to maintain precision
		_, err := io.WriteString(w, strconv.Quote(d.String()))
		if err != nil {
			panic(err)
		}
	})
}

// UnmarshalDecimal deserializes a GraphQL value to decimal.Decimal.
// Accepts both string and numeric inputs for flexibility.
func UnmarshalDecimal(v interface{}) (decimal.Decimal, error) {
	switch v := v.(type) {
	case string:
		// Parse from string (preferred for precision)
		d, err := decimal.NewFromString(v)
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid decimal string: %w", err)
		}

		return d, nil
	case int:
		// Accept integer input
		return decimal.NewFromInt(int64(v)), nil
	case int32:
		// Accept int32 input
		return decimal.NewFromInt(int64(v)), nil
	case int64:
		// Accept int64 input
		return decimal.NewFromInt(v), nil
	case float64:
		// Accept float input (may lose precision, but convenient for clients)
		return decimal.NewFromFloat(v), nil
	case float32:
		// Accept float32 input
		return decimal.NewFromFloat(float64(v)), nil
	default:
		return decimal.Decimal{}, fmt.Errorf("invalid type for Decimal: %T", v)
	}
}
