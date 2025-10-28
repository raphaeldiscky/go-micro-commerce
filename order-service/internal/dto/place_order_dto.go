// Package dto contains data transfer objects for order service.
package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PlaceOrderRequest represents the request to place an order from a checkout session.
type PlaceOrderRequest struct {
	CustomerID        uuid.UUID `json:"-"`                                       // from context or header
	CustomerEmail     string    `json:"-"`                                       // from context or header
	CheckoutSessionID uuid.UUID `json:"checkout_session_id" validate:"required"` // from cart-service
	IdempotencyKey    uuid.UUID `json:"idempotency_key"     validate:"required"` // generated from client
}

// PlaceOrderResponse represents the response when placing an order.
// Contains order details and payment gateway metadata for frontend payment processing.
type PlaceOrderResponse struct {
	Order           *OrderResponse  `json:"order"`
	PaymentMetadata PaymentMetadata `json:"payment_metadata"`
}

// PaymentMetadata contains payment gateway information needed for frontend payment processing.
type PaymentMetadata struct {
	PaymentID            uuid.UUID               `json:"payment_id"`
	PaymentGateway       constant.PaymentGateway `json:"payment_gateway"`
	GatewayTransactionID string                  `json:"gateway_transaction_id"` // e.g., pi_xxx for Stripe
	GatewayMetadata      map[string]any          `json:"gateway_metadata"`       // Contains client_secret, etc.
	Amount               decimal.Decimal         `json:"amount"`
	Currency             string                  `json:"currency"`
}
