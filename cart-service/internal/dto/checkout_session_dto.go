// Package dto contains data transfer objects for checkout session service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// Courier represents courier information.
type Courier struct {
	CourierID string `json:"courier_id"`
}

// Origin represents origin address information.
type Origin struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Destination represents destination address information.
type Destination struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Package represents package information.
type Package struct {
	WeightKG decimal.Decimal `json:"weight_kg"`
	Width    decimal.Decimal `json:"width"`
	Height   decimal.Decimal `json:"height"`
	Length   decimal.Decimal `json:"length"`
	Unit     string          `json:"unit"`
}

// CreateCheckoutSessionRequest represents the request to create a new checkout session.
type CreateCheckoutSessionRequest struct {
	CustomerID     uuid.UUID // from context or header
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required"`
	CartID         uuid.UUID `json:"cart_id"         validate:"required"`
}

// UpdateCheckoutSessionRequest represents the request to update a checkout session.
type UpdateCheckoutSessionRequest struct {
	CustomerID     uuid.UUID    // from context or header
	Courier        *Courier     `json:"courier,omitempty"`
	Destination    *Destination `json:"destination,omitempty"`
	Origin         *Origin      `json:"origin,omitempty"`
	Package        *Package     `json:"package,omitempty"`
	PaymentGateway *string      `json:"payment_gateway,omitempty"`
}

// PlaceOrderRequest represents the request to place an order from a checkout session.
type PlaceOrderRequest struct {
	CustomerID     uuid.UUID   `json:"-"` // from context or header
	IdempotencyKey uuid.UUID   `json:"idempotency_key" validate:"required"`
	Courier        Courier     `json:"courier"         validate:"required"`
	Destination    Destination `json:"destination"     validate:"required"`
	Origin         Origin      `json:"origin"          validate:"required"`
	Package        Package     `json:"package"         validate:"required"`
	PaymentGateway string      `json:"payment_gateway" validate:"required"`
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
	Courier        Courier                        `json:"courier"`
	Destination    Destination                    `json:"destination"`
	Origin         Origin                         `json:"origin"`
	Package        Package                        `json:"package"`
	Status         constant.CheckoutSessionStatus `json:"status"`
	PaymentGateway *string                        `json:"payment_gateway,omitempty"`
	Currency       string                         `json:"currency"`
	Items          []CheckoutSessionItemResponse  `json:"items"`
	CreatedAt      time.Time                      `json:"created_at"`
	UpdatedAt      time.Time                      `json:"updated_at"`
}
