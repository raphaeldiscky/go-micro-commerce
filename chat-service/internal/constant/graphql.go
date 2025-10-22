package constant

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// MarshalGQL implements graphql.Marshaler interface for MessageType.
func (m MessageType) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(m))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for MessageType.
func (m *MessageType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*m = MessageType(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for PresenceStatus.
func (p PresenceStatus) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(p))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for PresenceStatus.
func (p *PresenceStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*p = PresenceStatus(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for ConversationStatus.
func (c ConversationStatus) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(c))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for ConversationStatus.
func (c *ConversationStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*c = ConversationStatus(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for UserType.
func (u UserType) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(u))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for UserType.
func (u *UserType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*u = UserType(strings.ToLower(str))

	return nil
}

// MarshalGQL implements graphql.Marshaler interface for ParticipantRole.
func (p ParticipantRole) MarshalGQL(w io.Writer) {
	// Convert lowercase Go value to uppercase GraphQL value
	uppercase := strings.ToUpper(string(p))
	if _, err := fmt.Fprintf(w, "%q", uppercase); err != nil {
		panic(err)
	}
}

// UnmarshalGQL implements graphql.Unmarshaler interface for ParticipantRole.
func (p *ParticipantRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errors.New("enums must be strings")
	}
	// Convert uppercase GraphQL value to lowercase Go value
	*p = ParticipantRole(strings.ToLower(str))

	return nil
}
