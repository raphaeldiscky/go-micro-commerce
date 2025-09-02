package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
)

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata kafka.Metadata      `json:"metadata"`
	Payload  UserVerifiedPayload `json:"payload"`
}
