package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// DigitalWallet represents a digital wallet payment method.
type DigitalWallet struct {
	Type  constant.DigitalWalletType `json:"type"` // apple_pay, google_pay, paypal
	Token string                     `json:"token"`
	Email string                     `json:"email,omitempty"`
}

// Address represents a billing/shipping address.
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// PaymentGatewayRequest represents a payment processing request.
type PaymentGatewayRequest struct {
	TransactionID  uuid.UUID              `json:"transaction_id"`
	Amount         decimal.Decimal        `json:"amount"`
	Currency       string                 `json:"currency"`
	PaymentMethod  constant.PaymentMethod `json:"payment_method"`
	Card           *PaymentCard           `json:"card,omitempty"`
	DigitalWallet  *DigitalWallet         `json:"digital_wallet,omitempty"`
	BankAccount    *BankAccount           `json:"bank_account,omitempty"`
	Description    string                 `json:"description,omitempty"`
	CustomerID     uuid.UUID              `json:"customer_id"`
	CustomerEmail  string                 `json:"customer_email"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
}

// PaymentGatewayResponse represents the result of a payment processing.
type PaymentGatewayResponse struct {
	TransactionID   uuid.UUID                     `json:"transaction_id"`
	GatewayID       string                        `json:"gateway_id"`
	Status          constant.PaymentGatewayStatus `json:"status"`
	Amount          decimal.Decimal               `json:"amount"`
	Currency        string                        `json:"currency"`
	ProcessedAt     time.Time                     `json:"processed_at"`
	Fees            *decimal.Decimal              `json:"fees,omitempty"`
	NetworkFees     *decimal.Decimal              `json:"network_fees,omitempty"`
	GatewayResponse map[string]interface{}        `json:"gateway_response,omitempty"`
	FailureReason   string                        `json:"failure_reason,omitempty"`
	RequiresAction  bool                          `json:"requires_action,omitempty"`
	NextAction      *PaymentAction                `json:"next_action,omitempty"`
}

// PaymentAction represents an action required to complete payment.
type PaymentAction struct {
	Type constant.PaymentActionType `json:"type"`
	URL  string                     `json:"url,omitempty"`
	Data map[string]interface{}     `json:"data,omitempty"`
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	RefundID      uuid.UUID       `json:"refund_id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
	GatewayID     string          `json:"gateway_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Reason        string          `json:"reason,omitempty"`
}

// RefundResponse represents the result of a refund.
type RefundResponse struct {
	RefundID        uuid.UUID             `json:"refund_id"`
	TransactionID   uuid.UUID             `json:"transaction_id"`
	GatewayRefundID string                `json:"gateway_refund_id"`
	Status          constant.RefundStatus `json:"status"`
	Amount          decimal.Decimal       `json:"amount"`
	Currency        string                `json:"currency"`
	ProcessedAt     time.Time             `json:"processed_at"`
	Fees            *decimal.Decimal      `json:"fees,omitempty"`
}
