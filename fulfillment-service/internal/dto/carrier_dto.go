package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// ShippingRate represents the cost and estimated delivery for a shipping option.
type ShippingRate struct {
	CourierID          constant.CourierID `json:"courier_id"`
	CourierServiceName string             `json:"courier_service_name"`
	ShippingCost       decimal.Decimal    `json:"shipping_cost"`
	Currency           string             `json:"currency"`
	EstimatedDelivery  time.Time          `json:"estimated_delivery"`
	TransitDays        int                `json:"transit_days"`
}

// ShippingLabel represents a shipping label created by a courier.
type ShippingLabel struct {
	TrackingNumber     string             `json:"tracking_number"`
	LabelURL           string             `json:"label_url"`
	CourierID          constant.CourierID `json:"courier_id"`
	CourierServiceName string             `json:"courier_service_name"`
}

// TrackingInfo represents the current status of a shipment.
type TrackingInfo struct {
	TrackingNumber string                     `json:"tracking_number"`
	Status         constant.FulfillmentStatus `json:"status"`
	CourierID      constant.CourierID         `json:"courier_id"`
	LastUpdate     time.Time                  `json:"last_update"`
	Location       string                     `json:"location,omitempty"`
	Description    string                     `json:"description,omitempty"`
	DeliveredAt    *time.Time                 `json:"delivered_at,omitempty"`
}

// ShippingRequest represents a request to create a shipping label.
type ShippingRequest struct {
	OrderID     uuid.UUID          `json:"order_id"`
	CourierID   constant.CourierID `json:"courier_id"`
	Destination entity.Destination `json:"destination"`
	Origin      entity.Origin      `json:"origin"`
	Package     entity.Package     `json:"package"`
}
