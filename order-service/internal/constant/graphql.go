package constant

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// MarshalGQL implements graphql.Marshaler interface for OrderStatus.
func (s OrderStatus) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(s))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for OrderStatus.
func (s *OrderStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*s = OrderStatus(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for PaymentGateway.
func (p PaymentGateway) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(p))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for PaymentGateway.
func (p *PaymentGateway) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*p = PaymentGateway(strings.ToLower(str))

	return nil
}
