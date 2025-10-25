package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// Address represents a billing/shipping address (non-sensitive data, safe to transmit).
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// PaymentGatewayRequest represents a payment processing request.
// Uses PaymentMethod ID (pm_xxx) tokenized client-side for PCI DSS compliance.
// No raw card data is ever transmitted through this struct.
type PaymentGatewayRequest struct {
	PaymentMethodID string            `json:"payment_method_id"` // Stripe PM ID (pm_xxx) from client
	Metadata        map[string]string `json:"metadata,omitempty"`
	Amount          decimal.Decimal   `json:"amount"`
	Currency        string            `json:"currency"`
	Description     string            `json:"description,omitempty"`
	CustomerEmail   string            `json:"customer_email"`
	IdempotencyKey  string            `json:"idempotency_key"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"` // 24-hour payment window expiry
	TransactionID   uuid.UUID         `json:"transaction_id"`
	CustomerID      uuid.UUID         `json:"customer_id"`
}

// PaymentGatewayResponse represents the result of a payment processing.
// Includes client_secret for frontend to complete payment with Stripe.js.
type PaymentGatewayResponse struct {
	ProcessedAt     time.Time                     `json:"processed_at"`
	Fees            *decimal.Decimal              `json:"fees,omitempty"`
	NetworkFees     *decimal.Decimal              `json:"network_fees,omitempty"`
	GatewayResponse map[string]any                `json:"gateway_response,omitempty"`
	NextAction      *PaymentAction                `json:"next_action,omitempty"`
	ClientSecret    *string                       `json:"client_secret,omitempty"` // For stripe.confirmCardPayment()
	GatewayID       string                        `json:"gateway_id"`              // PaymentIntent ID (pi_xxx)
	Status          constant.PaymentGatewayStatus `json:"status"`
	Amount          decimal.Decimal               `json:"amount"`
	Currency        string                        `json:"currency"`
	FailureReason   string                        `json:"failure_reason,omitempty"`
	TransactionID   uuid.UUID                     `json:"transaction_id"`
	RequiresAction  bool                          `json:"requires_action,omitempty"` // Indicates 3DS needed
}

// PaymentAction represents an action required to complete payment.
type PaymentAction struct {
	Data map[string]any             `json:"data,omitempty"`
	Type constant.PaymentActionType `json:"type"`
	URL  string                     `json:"url,omitempty"`
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	GatewayID     string          `json:"gateway_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Reason        string          `json:"reason,omitempty"`
	RefundID      uuid.UUID       `json:"refund_id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
}

// RefundResponse represents the result of a refund.
type RefundResponse struct {
	ProcessedAt     time.Time             `json:"processed_at"`
	Fees            *decimal.Decimal      `json:"fees,omitempty"`
	GatewayRefundID string                `json:"gateway_refund_id"`
	Status          constant.RefundStatus `json:"status"`
	Amount          decimal.Decimal       `json:"amount"`
	Currency        string                `json:"currency"`
	RefundID        uuid.UUID             `json:"refund_id"`
	TransactionID   uuid.UUID             `json:"transaction_id"`
}

// SetupIntentRequest represents a request to create a SetupIntent for collecting payment method.
// Used for delayed payment confirmation pattern (save now, charge later).
type SetupIntentRequest struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerEmail string    `json:"customer_email"`
	OrderID       uuid.UUID `json:"order_id"`
}

// ChargeOffSessionRequest represents a request to charge a saved payment method without customer present.
// Used for delayed payment confirmation when customer already provided payment details.
type ChargeOffSessionRequest struct {
	PaymentMethodID  string          `json:"payment_method_id"`  // Stripe PM ID (pm_xxx)
	StripeCustomerID string          `json:"stripe_customer_id"` // Stripe Customer ID (cus_xxx)
	Amount           decimal.Decimal `json:"amount"`
	Currency         string          `json:"currency"`
	TransactionID    uuid.UUID       `json:"transaction_id"`
	OrderID          uuid.UUID       `json:"order_id"`
	Description      string          `json:"description"`
}
