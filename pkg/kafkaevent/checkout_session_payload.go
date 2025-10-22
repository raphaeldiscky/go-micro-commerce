package kafkaevent

import (
	"time"

	"github.com/google/uuid"
)

// CheckoutSessionOrderPlacedPayload holds the data for the checkout session confirmed event (place order).
type CheckoutSessionOrderPlacedPayload struct {
	CheckoutSessionID uuid.UUID             `json:"checkout_session_id"`
	IdempotencyKey    uuid.UUID             `json:"idempotency_key"`
	UserID            uuid.UUID             `json:"user_id"`
	Items             []CheckoutItemPayload `json:"items"`
	Status            string                `json:"status"`
	Currency          string                `json:"currency"`
	PaymentGateway    string                `json:"payment_gateway"`
	PaymentMethod     string                `json:"payment_method"`
	CreatedAt         time.Time             `json:"created_at"`
}

// CheckoutItemPayload holds the data for the checkout session item.
type CheckoutItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

// CheckoutSessionCanceledPayload holds the data for the checkout session canceled event.
type CheckoutSessionCanceledPayload struct {
	CheckoutSessionID uuid.UUID `json:"checkout_session_id"`
	IdempotencyKey    uuid.UUID `json:"idempotency_key"`
	Reason            string    `json:"reason"`
	CreatedAt         time.Time `json:"created_at"`
}
