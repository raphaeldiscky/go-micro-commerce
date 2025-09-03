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
	OrderID             uuid.UUID       `json:"order_id"              validate:"required"`
	TrackingNumber      string          `json:"tracking_number"       validate:"required"`
	ShippingCost        decimal.Decimal `json:"shipping_cost"         validate:"gte=0"`
	Currency            string          `json:"currency"              validate:"required,len=3"`
	WeightKG            decimal.Decimal `json:"weight_kg"             validate:"required,gt=0"`
	EstimatedDeliveryAt time.Time       `json:"estimated_delivery_at" validate:"required"`
}

// UpdateFulfillmentStatusRequest represents the request to update fulfillment status.
type UpdateFulfillmentStatusRequest struct {
	Status constant.FulfillmentStatus `json:"status" validate:"required"`
}

// SetCarrierInfoRequest represents the request to set carrier information.
type SetCarrierInfoRequest struct {
	Carrier          string  `json:"carrier"            validate:"required"`
	ShippingLabelURL *string `json:"shipping_label_url"`
}

// SetDimensionsRequest represents the request to set package dimensions.
type SetDimensionsRequest struct {
	Dimensions entity.Dimensions `json:"dimensions" validate:"required"`
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
	Carrier             *string                    `json:"carrier,omitempty"`
	ShippingLabelURL    *string                    `json:"shipping_label_url,omitempty"`
	ShippingCost        decimal.Decimal            `json:"shipping_cost"`
	Currency            string                     `json:"currency"`
	WeightKG            decimal.Decimal            `json:"weight_kg"`
	Dimensions          *entity.Dimensions         `json:"dimensions,omitempty"`
	EstimatedDeliveryAt time.Time                  `json:"estimated_delivery_at"`
	ActualDeliveryAt    *time.Time                 `json:"actual_delivery_at,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

// GetShippingRatesRequest represents the request to get shipping rates.
type GetShippingRatesRequest struct {
	FromAddress ShippingAddress `json:"from_address" validate:"required"`
	ToAddress   ShippingAddress `json:"to_address"   validate:"required"`
	Package     Package         `json:"package"      validate:"required"`
	OrderID     uuid.UUID       `json:"order_id"     validate:"required"`
}

// ShippingRateResponse represents a shipping rate option.
type ShippingRateResponse struct {
	Carrier           constant.CarrierType `json:"carrier"`
	Service           string               `json:"service"`
	Cost              decimal.Decimal      `json:"cost"`
	Currency          string               `json:"currency"`
	EstimatedDelivery time.Time            `json:"estimated_delivery"`
	TransitDays       int                  `json:"transit_days"`
}

// CreateShipmentRequest represents the request to create a shipment.
type CreateShipmentRequest struct {
	Carrier         string           `json:"carrier"                    validate:"required"`
	Service         string           `json:"service"                    validate:"required"`
	FromAddress     ShippingAddress  `json:"from_address"               validate:"required"`
	ToAddress       ShippingAddress  `json:"to_address"                 validate:"required"`
	Package         Package          `json:"package"                    validate:"required"`
	InsuranceAmount *decimal.Decimal `json:"insurance_amount,omitempty"`
	Signature       bool             `json:"signature,omitempty"`
}
