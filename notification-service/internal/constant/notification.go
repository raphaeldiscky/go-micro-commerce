package constant

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// NotificationType represents the type of notification.
//
//nolint:recvcheck // Mixed receivers required: Unmarshal uses pointer, Marshal uses value.
type NotificationType string

// Notification type constants.
const (
	NotificationTypeNewMessage     NotificationType = "new_message"
	NotificationTypeNewProduct     NotificationType = "new_product"
	NotificationTypeOrderUpdate    NotificationType = "order_update"
	NotificationTypeOrderConfirmed NotificationType = "order_confirmed"
	NotificationTypeOrderShipped   NotificationType = "order_shipped"
	NotificationTypeOrderDelivered NotificationType = "order_delivered"
	NotificationTypeOrderCancelled NotificationType = "order_cancelled"
	NotificationTypePaymentSuccess NotificationType = "payment_success"
	NotificationTypeSystemAlert    NotificationType = "system_alert"
)

// IsValid checks if the notification type is valid.
func (e NotificationType) IsValid() bool {
	switch e {
	case NotificationTypeNewMessage,
		NotificationTypeNewProduct,
		NotificationTypeOrderUpdate,
		NotificationTypeOrderConfirmed,
		NotificationTypeOrderShipped,
		NotificationTypeOrderDelivered,
		NotificationTypeOrderCancelled,
		NotificationTypePaymentSuccess,
		NotificationTypeSystemAlert:
		return true
	}

	return false
}

// String returns the string representation (lowercase with underscore).
func (e NotificationType) String() string {
	return string(e)
}

// ToGraphQL converts lowercase database format to uppercase GraphQL format.
// Example: "system_alert" -> "SYSTEM_ALERT".
func (e NotificationType) ToGraphQL() string {
	return strings.ToUpper(string(e))
}

// FromGraphQL converts uppercase GraphQL format to lowercase database format.
// Example: "SYSTEM_ALERT" -> "system_alert".
func FromGraphQL(s string) NotificationType {
	return NotificationType(strings.ToLower(s))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface.
// Converts GraphQL enum (uppercase) to Go constant (lowercase).
func (e *NotificationType) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}

	*e = FromGraphQL(str)

	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid NotificationType", str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface.
// Converts Go constant (lowercase) to GraphQL enum (uppercase).
func (e NotificationType) MarshalGQL(w io.Writer) {
	//nolint:errcheck // Writer errors are not actionable in marshaling context.
	fmt.Fprint(w, strconv.Quote(e.ToGraphQL()))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *NotificationType) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	return e.UnmarshalGQL(s)
}

// MarshalJSON implements the json.Marshaler interface.
func (e NotificationType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	e.MarshalGQL(&buf)

	return buf.Bytes(), nil
}
