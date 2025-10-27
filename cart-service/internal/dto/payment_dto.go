// Package dto provides data transfer objects for payment operations.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreatePaymentIntentRequest represents the request to create a PaymentIntent.
type CreatePaymentIntentRequest struct {
	OrderID           uuid.UUID        `json:"order_id"`
	CustomerID        uuid.UUID        `json:"customer_id"`
	CustomerEmail     string           `json:"customer_email"`
	Amount            decimal.Decimal  `json:"amount"`
	Currency          string           `json:"currency"`
	PaymentGateway    string           `json:"payment_gateway"`
	CheckoutSessionID uuid.UUID        `json:"checkout_session_id"`
	IdempotencyKey    uuid.UUID        `json:"idempotency_key"`
	Items             []PaymentItemDTO `json:"items"`
}

// PaymentItemDTO represents a payment item for PaymentIntent creation.
type PaymentItemDTO struct {
	ProductID   uuid.UUID       `json:"product_id"`
	ProductName string          `json:"product_name"`
	Quantity    int64           `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Currency    string          `json:"currency"`
}

// CreatePaymentIntentResponse represents the response from creating a PaymentIntent.
type CreatePaymentIntentResponse struct {
	PaymentIntentID string     `json:"payment_intent_id"`
	ClientSecret    string     `json:"client_secret"`
	PaymentGateway  string     `json:"payment_gateway"`
	Status          string     `json:"status"`
	Amount          string     `json:"amount"`
	Currency        string     `json:"currency"`
	OrderID         string     `json:"order_id"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}
