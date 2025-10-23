package constant

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// MarshalGQL implements graphql.Marshaler interface for CheckoutSessionStatus.
func (s CheckoutSessionStatus) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(s))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for CheckoutSessionStatus.
func (s *CheckoutSessionStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*s = CheckoutSessionStatus(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for CartStatus.
func (s CartStatus) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(s))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for CartStatus.
func (s *CartStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*s = CartStatus(strings.ToLower(str))

	return nil
}
