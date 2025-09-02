// Package dto provides data transfer objects for the payment service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// CreatePaymentRequest represents the request to create a payment from an order event.
type CreatePaymentRequest struct {
	OrderID       uuid.UUID              `json:"order_id"       validate:"required"`
	Amount        decimal.Decimal        `json:"amount"         validate:"required,gt=0"`
	Currency      string                 `json:"currency"       validate:"required,len=3"`
	PaymentMethod constant.PaymentMethod `json:"payment_method" validate:"required"`
}

// ProcessPaymentRequest represents the request to process a payment.
type ProcessPaymentRequest struct {
	CustomerID     uuid.UUID              `json:"customer_id"`
	CustomerEmail  string                 `json:"customer_email"`
	IdempotencyKey uuid.UUID              `json:"idempotency_key" validate:"required"`
	PaymentMethod  constant.PaymentMethod `json:"payment_method"  validate:"required"`
}

// PaymentResponse represents the response for payment operations.
type PaymentResponse struct {
	ID                 uuid.UUID              `json:"id"`
	OrderID            uuid.UUID              `json:"order_id"`
	Amount             decimal.Decimal        `json:"amount"`
	Currency           string                 `json:"currency"`
	Status             constant.PaymentStatus `json:"status"`
	PaymentMethod      constant.PaymentMethod `json:"payment_method"`
	PaymentGateway     *string                `json:"payment_gateway,omitempty"`
	GatewayReferenceID *string                `json:"gateway_reference_id,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	FailedAt           *time.Time             `json:"failed_at,omitempty"`
}
