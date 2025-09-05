package event

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FulfillmentItemPayload holds the data for each item in the fulfillment.
type FulfillmentItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

// FulfillmentRequestPayload holds the data for the Fulfillment Request event.
type FulfillmentRequestPayload struct {
	OrderID         uuid.UUID                `json:"order_id"`
	CustomerID      uuid.UUID                `json:"customer_id"`
	TotalPrice      decimal.Decimal          `json:"total_price"`
	Currency        string                   `json:"currency"`
	Items           []FulfillmentItemPayload `json:"items"`
	ShippingAddress ShippingAddressPayload   `json:"shipping_address"`
}

// ShippingAddressPayload holds the shipping address information.
type ShippingAddressPayload struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// FulfillmentLifecyclePayload holds the data for the Fulfillment Lifecycle event.
type FulfillmentLifecyclePayload struct {
	FulfillmentID     uuid.UUID `json:"fulfillment_id"`
	OrderID           uuid.UUID `json:"order_id"`
	Status            string    `json:"status"`
	TrackingNumber    string    `json:"tracking_number,omitempty"`
	EstimatedDelivery string    `json:"estimated_delivery,omitempty"`
}
