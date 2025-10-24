// Package dto contains data transfer objects for checkout session service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// CreateCheckoutSessionRequest represents the request to create a new checkout session.
type CreateCheckoutSessionRequest struct {
	CustomerID     uuid.UUID // from context or header
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required"`
	CartID         uuid.UUID `json:"cart_id"         validate:"required"`
}

// PlaceOrderRequest represents the request to place an order from a checkout session.
type PlaceOrderRequest struct {
	CustomerID     uuid.UUID // from context or header
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required"`
	AddressID      uuid.UUID `json:"address_id"      validate:"required"`
	CarrierID      string    `json:"carrier_id"      validate:"required"`
	PaymentMethod  string    `json:"payment_method"  validate:"required"`
	PaymentGateway string    `json:"payment_gateway" validate:"required"`
}

// CheckoutSessionItemResponse represents a checkout session item in API responses.
type CheckoutSessionItemResponse struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"product_id"`
	ProductName string          `json:"product_name"`
	Quantity    int64           `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
}

// CheckoutSessionResponse represents a checkout session in API responses.
type CheckoutSessionResponse struct {
	ID             uuid.UUID                      `json:"id"`
	IdempotencyKey uuid.UUID                      `json:"idempotency_key"`
	CustomerID     uuid.UUID                      `json:"customer_id"`
	AddressID      *uuid.UUID                     `json:"address_id,omitempty"`
	CarrierID      *string                        `json:"carrier_id,omitempty"`
	Status         constant.CheckoutSessionStatus `json:"status"`
	PaymentGateway *string                        `json:"payment_gateway,omitempty"`
	PaymentMethod  *string                        `json:"payment_method,omitempty"`
	Currency       string                         `json:"currency"`
	Items          []CheckoutSessionItemResponse  `json:"items"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
}

// UpdatePaymentRequest represents the request to update payment method for a checkout session.
type UpdatePaymentRequest struct {
	CustomerID        uuid.UUID // from context or header
	CheckoutSessionID uuid.UUID // from URL param
	PaymentMethod     string    `json:"payment_method"  validate:"required"`
	PaymentGateway    string    `json:"payment_gateway" validate:"required"`
}

// UpdatePaymentResponse represents the response to update payment method for a checkout session.
type UpdatePaymentResponse struct {
	PaymentGateway constant.PaymentGateway `json:"payment_gateway"`
	PaymentMethod  constant.PaymentMethod  `json:"payment_method"`
}

// UpdateCarrierRequest represents the request to update carrier for a checkout session.
type UpdateCarrierRequest struct {
	CustomerID        uuid.UUID // from context or header
	CheckoutSessionID uuid.UUID // from URL param
	CarrierID         string    `json:"carrier_id" validate:"required"`
}

// UpdateCarrierResponse represents the response to update carrier for a checkout session.
type UpdateCarrierResponse struct {
	ShippingCost decimal.Decimal `json:"shipping_cost"`
}
