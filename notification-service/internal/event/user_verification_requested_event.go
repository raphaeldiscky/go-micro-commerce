package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata mq.KafkaMetadata                  `json:"metadata"`
	Payload  EmailVerificationRequestedPayload `json:"payload"`
}
