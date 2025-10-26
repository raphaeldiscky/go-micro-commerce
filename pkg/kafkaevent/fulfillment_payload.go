package kafkaevent

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FulfillmentItemPayload holds the data for each item in the fulfillment.
type FulfillmentItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

// Package represents the package dimensions and weight in kg.
type Package struct {
	WeightKG decimal.Decimal `json:"weight_kg"`
	Width    decimal.Decimal `json:"width"`
	Height   decimal.Decimal `json:"height"`
	Length   decimal.Decimal `json:"length"`
	Unit     string          `json:"unit"`
}

// FulfillmentRequestPayload holds the data for the Fulfillment Request event.
type FulfillmentRequestPayload struct {
	OrderID     uuid.UUID                `json:"order_id"`
	CustomerID  uuid.UUID                `json:"customer_id"`
	Currency    string                   `json:"currency"`
	Items       []FulfillmentItemPayload `json:"items"`
	Courier     Courier                  `json:"courier"`
	Destination Destination              `json:"destination"`
	Origin      Origin                   `json:"origin"`
	Package     Package                  `json:"package"`
}

// Courier represents the courier information.
type Courier struct {
	CourierID string `json:"courier_id"`
}

// Destination holds the shipping address information.
type Destination struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Origin holds the warehouse address information.
type Origin struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// FulfillmentLifecyclePayload holds the data for the Fulfillment Lifecycle event.
type FulfillmentLifecyclePayload struct {
	FulfillmentID       uuid.UUID       `json:"fulfillment_id"`
	OrderID             uuid.UUID       `json:"order_id"`
	Status              string          `json:"status"`
	ShippingCost        decimal.Decimal `json:"shipping_cost,omitempty"`
	TrackingNumber      string          `json:"tracking_number,omitempty"`
	EstimatedDeliveryAt time.Time       `json:"estimated_delivery_at,omitempty"`
}
