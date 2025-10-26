// Package dto provides data transfer objects for the fulfillment service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// CreateFulfillmentRequest represents the request to create a fulfillment from an order event.
type CreateFulfillmentRequest struct {
	OrderID             uuid.UUID          `json:"order_id"              validate:"required"`
	TrackingNumber      string             `json:"tracking_number"       validate:"required"`
	ShippingCost        decimal.Decimal    `json:"shipping_cost"         validate:"gte=0"`
	Currency            string             `json:"currency"              validate:"required,len=3"`
	Origin              entity.Origin      `json:"origin"                validate:"required"`
	Destination         entity.Destination `json:"destination"           validate:"required"`
	Package             entity.Package     `json:"package"               validate:"required"`
	EstimatedDeliveryAt time.Time          `json:"estimated_delivery_at" validate:"required"`
}

// UpdateFulfillmentStatusRequest represents the request to update fulfillment status.
type UpdateFulfillmentStatusRequest struct {
	Status constant.FulfillmentStatus `json:"status" validate:"required"`
}

// SetCourierInfoRequest represents the request to set Courier information.
type SetCourierInfoRequest struct {
	CourierID        constant.CourierID `json:"courier_id"         validate:"required"`
	ShippingLabelURL string             `json:"shipping_label_url" validate:"required"`
}

// SetPackageRequest represents the request to set package.
type SetPackageRequest struct {
	Package entity.Package `json:"package" validate:"required"`
}

// SetActualDeliveryRequest represents the request to set actual delivery time.
type SetActualDeliveryRequest struct {
	ActualDeliveryAt time.Time `json:"actual_delivery_at" validate:"required"`
}

// FulfillmentResponse represents the response for fulfillment operations.
type FulfillmentResponse struct {
	ID                  uuid.UUID                  `json:"id"`
	OrderID             uuid.UUID                  `json:"order_id"`
	Status              constant.FulfillmentStatus `json:"status"`
	TrackingNumber      string                     `json:"tracking_number"`
	CourierID           constant.CourierID         `json:"Courier,omitempty"`
	ShippingLabelURL    string                     `json:"shipping_label_url,omitempty"`
	ShippingCost        decimal.Decimal            `json:"shipping_cost"`
	Currency            string                     `json:"currency"`
	Origin              entity.Origin              `json:"origin"`
	Destination         entity.Destination         `json:"destination"`
	Package             entity.Package             `json:"Package,omitempty"`
	EstimatedDeliveryAt time.Time                  `json:"estimated_delivery_at"`
	ActualDeliveryAt    *time.Time                 `json:"actual_delivery_at,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

// CalculateShippingRatesRequest represents the request to get many shipping rates.
type CalculateShippingRatesRequest struct {
	Currency    string             `json:"currency"    validate:"required,len=3"`
	CourierID   constant.CourierID `json:"courier_id"  validate:"required"`
	Destination entity.Destination `json:"destination" validate:"required"`
	Origin      entity.Origin      `json:"origin"      validate:"required"`
	Package     entity.Package     `json:"Package"     validate:"required"` // weight_kg, width, height, length in cm
}

// CalculateShippingRateRequest represents the request to get single shipping rates.
type CalculateShippingRateRequest struct {
	Currency    string             `json:"currency"    validate:"required,len=3"`
	CourierID   constant.CourierID `json:"courier_id"  validate:"required"`
	Destination entity.Destination `json:"destination" validate:"required"`
	Origin      entity.Origin      `json:"origin"      validate:"required"`
	Package     entity.Package     `json:"Package"     validate:"required"` // weight_kg, width, height, length in cm
}

// ShippingRateResponse represents a shipping rate option.
type ShippingRateResponse struct {
	CourierID          constant.CourierID `json:"courier_id"`
	CourierServiceName string             `json:"courier_service_name"`
	ShippingCost       decimal.Decimal    `json:"shipping_cost"`
	Currency           string             `json:"currency"`
	EstimatedDelivery  time.Time          `json:"estimated_delivery"`
	TransitDays        int                `json:"transit_days"`
}

// CreateShipmentRequest represents the request to create a shipment.
type CreateShipmentRequest struct {
	CourierID   constant.CourierID `json:"courier_id"  validate:"required"`
	Destination entity.Destination `json:"destination" validate:"required"`
	Origin      entity.Origin      `json:"origin"      validate:"required"`
	Package     entity.Package     `json:"Package"     validate:"required"` // weight_kg, width, height, length in cm
}
