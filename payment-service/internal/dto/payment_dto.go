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
	Amount        decimal.Decimal        `json:"amount"         validate:"required,gt=0"`
	Currency      string                 `json:"currency"       validate:"required,len=3"`
	PaymentMethod constant.PaymentMethod `json:"payment_method" validate:"required"`
	OrderID       uuid.UUID              `json:"order_id"       validate:"required"`
}

// ProcessPaymentRequest represents the request to process a payment.
type ProcessPaymentRequest struct {
	CustomerEmail  string                 `json:"customer_email"`
	PaymentMethod  constant.PaymentMethod `json:"payment_method"  validate:"required"`
	CustomerID     uuid.UUID              `json:"customer_id"`
	IdempotencyKey uuid.UUID              `json:"idempotency_key" validate:"required"`
}

// PaymentResponse represents the response for payment operations.
type PaymentResponse struct {
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	PaymentGateway     *string                `json:"payment_gateway,omitempty"`
	GatewayReferenceID *string                `json:"gateway_reference_id,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	FailedAt           *time.Time             `json:"failed_at,omitempty"`
	Amount             decimal.Decimal        `json:"amount"`
	Currency           string                 `json:"currency"`
	Status             constant.PaymentStatus `json:"status"`
	PaymentMethod      constant.PaymentMethod `json:"payment_method"`
	ID                 uuid.UUID              `json:"id"`
	OrderID            uuid.UUID              `json:"order_id"`
}

// PaymentCard represents a payment card.
type PaymentCard struct {
	BillingAddress *Address `json:"billing_address,omitempty"`
	Number         string   `json:"number"`
	CVV            string   `json:"cvv"`
	HolderName     string   `json:"holder_name"`
	ExpiryMonth    int      `json:"expiry_month"`
	ExpiryYear     int      `json:"expiry_year"`
}
