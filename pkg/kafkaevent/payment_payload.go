package kafkaevent

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentRequestPayload holds the data for payment request events.
type PaymentRequestPayload struct {
	PaymentID      uuid.UUID       `json:"payment_id"`
	OrderID        uuid.UUID       `json:"order_id"`
	CustomerID     uuid.UUID       `json:"customer_id"`
	TotalPrice     decimal.Decimal `json:"total_price"`
	Currency       string          `json:"currency"`
	PaymentMethod  string          `json:"payment_method"`
	PaymentGateway string          `json:"payment_gateway"`
}

// PaymentRefundPayload holds the data for payment refund events.
type PaymentRefundPayload struct {
	OrderID    uuid.UUID       `json:"order_id"`
	CustomerID uuid.UUID       `json:"customer_id"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Reason     string          `json:"reason"`
	Timestamp  string          `json:"timestamp"`
}

// PaymentLifecyclePayload holds the data for the Payment Lifecycle event.
type PaymentLifecyclePayload struct {
	PaymentID    uuid.UUID       `json:"payment_id"`
	OrderID      uuid.UUID       `json:"order_id"`
	Status       string          `json:"status"`
	TotalPrice   decimal.Decimal `json:"total_price"`
	ClientSecret *string         `json:"client_secret,omitempty"` // Stripe client secret for Payment Element
	ExpiresAt    *time.Time      `json:"expires_at,omitempty"`    // 24-hour payment window expiry
}
