package event

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

// Dimensions represents the package dimensions.
type Dimensions struct {
	Width  decimal.Decimal `json:"width"`
	Height decimal.Decimal `json:"height"`
	Length decimal.Decimal `json:"length"`
	Unit   string          `json:"unit"`
}

// FulfillmentRequestPayload holds the data for the Fulfillment Request event.
type FulfillmentRequestPayload struct {
	OrderID    uuid.UUID                `json:"order_id"`
	CustomerID uuid.UUID                `json:"customer_id"`
	Currency   string                   `json:"currency"`
	Items      []FulfillmentItemPayload `json:"items"`
	Shipping   Shipping                 `json:"shipping"`
}

// Shipping represents the shipping data for an order.
type Shipping struct {
	CarrierID   string             `json:"carrier_id"`
	FromAddress FromAddressPayload `json:"from_address"`
	ToAddress   ToAddressPayload   `json:"to_address"`
	WeightKG    decimal.Decimal    `json:"weight_kg"`
	Dimensions  Dimensions         `json:"dimensions"`
}

// ToAddressPayload holds the shipping address information.
type ToAddressPayload struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// FromAddressPayload holds the warehouse address information.
type FromAddressPayload struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
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
