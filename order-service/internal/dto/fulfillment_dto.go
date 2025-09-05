package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FulfillmentResponse represents the response from fulfillment service.
type FulfillmentResponse struct {
	FulfillmentID  uuid.UUID       `json:"fulfillment_id"`
	TrackingNumber string          `json:"tracking_number"`
	ShippingCost   decimal.Decimal `json:"shipping_cost"`
	Status         string          `json:"status"`
	OrderID        uuid.UUID       `json:"order_id"`
	Error          error           `json:"error,omitempty"`
}
