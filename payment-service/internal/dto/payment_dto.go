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
	Amount           decimal.Decimal         `json:"amount"                       validate:"required,gt=0"`
	Currency         string                  `json:"currency"                     validate:"required,len=3"`
	PaymentGateway   constant.PaymentGateway `json:"payment_gateway"              validate:"required"`
	OrderID          uuid.UUID               `json:"order_id"                     validate:"required"`
	PaymentMethodID  string                  `json:"payment_method_id,omitempty"`  // Optional Stripe PM ID
	StripeCustomerID string                  `json:"stripe_customer_id,omitempty"` // Optional Stripe Customer ID
}

// ProcessPaymentRequest represents the request to process a payment.
// Uses Stripe PaymentMethod ID (pm_xxx) created client-side for PCI compliance.
type ProcessPaymentRequest struct {
	PaymentMethodID string    `json:"payment_method_id" validate:"required"` // Stripe PM ID (pm_xxx)
	CustomerEmail   string    `json:"customer_email"`
	CustomerID      uuid.UUID `json:"customer_id"`
	IdempotencyKey  uuid.UUID `json:"idempotency_key"   validate:"required"`
}

// PaymentResponse represents the response for payment operations.
type PaymentResponse struct {
	ID                 uuid.UUID               `json:"id"`
	OrderID            uuid.UUID               `json:"order_id"`
	Status             constant.PaymentStatus  `json:"status"`
	PaymentGateway     constant.PaymentGateway `json:"payment_gateway"`
	GatewayReferenceID *string                 `json:"gateway_reference_id,omitempty"`
	PaymentMethodID    *string                 `json:"payment_method_id,omitempty"`  // Stripe PM ID (pm_xxx)
	StripeCustomerID   *string                 `json:"stripe_customer_id,omitempty"` // Stripe Customer ID (cus_xxx)
	ClientSecret       *string                 `json:"client_secret,omitempty"`      // For stripe.confirmCardPayment()
	RequiresAction     bool                    `json:"requires_action"`              // Indicates 3DS/authentication needed
	NextActionType     *string                 `json:"next_action_type,omitempty"`   // Type of action required
	Amount             decimal.Decimal         `json:"amount"`
	Currency           string                  `json:"currency"`
	ExpiresAt          *time.Time              `json:"expires_at,omitempty"` // 24-hour payment window expiry
	CompletedAt        *time.Time              `json:"completed_at,omitempty"`
	FailedAt           *time.Time              `json:"failed_at,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// PaymentMethodInfo represents non-sensitive payment method information.
// Sensitive card data is tokenized client-side using Stripe.js for PCI compliance.
// This struct only contains data that is safe to transmit and store.
type PaymentMethodInfo struct {
	Type        string `json:"type"`         // card, apple_pay, google_pay, ideal, etc.
	Last4       string `json:"last4"`        // Last 4 digits (safe to store per PCI DSS)
	Brand       string `json:"brand"`        // visa, mastercard, amex, etc.
	ExpiryMonth int    `json:"expiry_month"` // Safe to store per PCI DSS
	ExpiryYear  int    `json:"expiry_year"`  // Safe to store per PCI DSS
	Country     string `json:"country"`      // Card issuing country
}
