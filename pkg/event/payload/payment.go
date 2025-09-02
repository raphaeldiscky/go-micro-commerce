package payload

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// PaymentRequestPayload holds the data for payment request events.
type PaymentRequestPayload struct {
	PaymentID     uuid.UUID       `json:"payment_id"`
	OrderID       uuid.UUID       `json:"order_id"`
	CustomerID    uuid.UUID       `json:"customer_id"`
	TotalPrice    decimal.Decimal `json:"total_price"`
	Currency      string          `json:"currency"`
	PaymentMethod string          `json:"payment_method"`
}
