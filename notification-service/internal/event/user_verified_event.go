package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata mq.KafkaMetadata    `json:"metadata"`
	Payload  UserVerifiedPayload `json:"payload"`
}
