package kafkaevent

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	GatewayMetadata   json.RawMessage       `json:"gateway_metadata"`
	ShippingCost      decimal.Decimal       `json:"shipping_cost"`
	TotalAmount       decimal.Decimal       `json:"total_amount"`
	Courier           Courier               `json:"courier"`
	Destination       Destination           `json:"destination"`
	Origin            Origin                `json:"origin"`
	Package           Package               `json:"package"`
	CreatedAt         time.Time             `json:"created_at"`
}

// CheckoutItemPayload holds the data for the checkout session item.
type CheckoutItemPayload struct {
	ProductID   uuid.UUID       `json:"product_id"`
	ProductName string          `json:"product_name"`
	Quantity    int64           `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
}

// CheckoutSessionCanceledPayload holds the data for the checkout session canceled event.
type CheckoutSessionCanceledPayload struct {
	CheckoutSessionID uuid.UUID `json:"checkout_session_id"`
	IdempotencyKey    uuid.UUID `json:"idempotency_key"`
	Reason            string    `json:"reason"`
	CreatedAt         time.Time `json:"created_at"`
}
