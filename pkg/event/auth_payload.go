package event

import (
	"time"

	"github.com/google/uuid"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	Token          string    `json:"token"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
}

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}
