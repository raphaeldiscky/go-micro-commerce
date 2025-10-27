package dto

import (
	"github.com/google/uuid"
)

// CheckoutSessionReminderRequest represents the request for a checkout session reminder task.
type CheckoutSessionReminderRequest struct {
	CheckoutSessionID uuid.UUID `json:"checkout_session_id"`
	CustomerEmail     string    `json:"customer_email"`
}
