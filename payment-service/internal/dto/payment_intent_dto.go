// Package dto provides data transfer objects for the payment service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// CreatePaymentIntentRequestDTO represents the request to create a payment intent.
// This is used for synchronous gRPC payment intent creation before order processing.
type CreatePaymentIntentRequestDTO struct {
	OrderID        uuid.UUID               `json:"order_id"        validate:"required"`
	Amount         decimal.Decimal         `json:"amount"          validate:"required,gt=0"`
	Currency       string                  `json:"currency"        validate:"required,len=3"`
	PaymentGateway constant.PaymentGateway `json:"payment_gateway" validate:"required"`
	CustomerEmail  string                  `json:"customer_email"  validate:"omitempty,email"`
	CustomerID     uuid.UUID               `json:"customer_id"     validate:"required"`
}

// CreatePaymentIntentResponseDTO represents the response when creating a payment intent.
// Contains all necessary data for the frontend to process payment.
type CreatePaymentIntentResponseDTO struct {
	PaymentID            uuid.UUID               `json:"payment_id"`
	OrderID              uuid.UUID               `json:"order_id"`
	Amount               decimal.Decimal         `json:"amount"`
	Currency             string                  `json:"currency"`
	PaymentGateway       constant.PaymentGateway `json:"payment_gateway"`
	GatewayTransactionID string                  `json:"gateway_transaction_id"` // e.g., pi_xxx for Stripe
	GatewayMetadata      map[string]any          `json:"gateway_metadata"`       // Contains client_secret, etc.
	ExpiresAt            *time.Time              `json:"expires_at,omitempty"`
	CreatedAt            time.Time               `json:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at"`
}
