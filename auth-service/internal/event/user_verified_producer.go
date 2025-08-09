package event

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/constant"
)

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata KafkaMetadata
	Payload  UserVerifiedPayload
}

// GetPayload returns the data associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// NewUserVerifiedEvent creates a new UserVerifiedEvent.
func NewUserVerifiedEvent(
	userID uuid.UUID,
	email string,
) *UserVerifiedEvent {
	return &UserVerifiedEvent{
		Metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeUserVerified,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceAuthService,
		},
		Payload: UserVerifiedPayload{
			UserID: userID,
			Email:  email,
		},
	}
}
