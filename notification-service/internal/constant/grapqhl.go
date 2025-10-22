package constant

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ToGraphQL converts lowercase database format to uppercase GraphQL format.
// Example: "system_alert" -> "SYSTEM_ALERT".
func (e PushNotificationType) ToGraphQL() string {
	return strings.ToUpper(string(e))
}

// FromGraphQL converts uppercase GraphQL format to lowercase database format.
// Example: "SYSTEM_ALERT" -> "system_alert".
func FromGraphQL(s string) PushNotificationType {
	return PushNotificationType(strings.ToLower(s))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface.
// Converts GraphQL enum (uppercase) to Go constant (lowercase).
func (e *PushNotificationType) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}

	*e = FromGraphQL(str)

	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PushNotificationType", str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface.
// Converts Go constant (lowercase) to GraphQL enum (uppercase).
func (e PushNotificationType) MarshalGQL(w io.Writer) {
	//nolint:errcheck // Writer errors are not actionable in marshaling context.
	fmt.Fprint(w, strconv.Quote(e.ToGraphQL()))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *PushNotificationType) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	return e.UnmarshalGQL(s)
}

// MarshalJSON implements the json.Marshaler interface.
func (e PushNotificationType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	e.MarshalGQL(&buf)

	return buf.Bytes(), nil
}
