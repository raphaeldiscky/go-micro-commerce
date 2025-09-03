// Package entity defines the Fulfillment entity and its validation logic.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// Dimensions represents the package dimensions.
type Dimensions struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Length float64 `json:"length"`
	Unit   string  `json:"unit"` // cm or inch
}

// Fulfillment represents a fulfillment record in the marketplace.
type Fulfillment struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID // Reference to order from order-service
	Status              constant.FulfillmentStatus
	TrackingNumber      string
	Carrier             *string
	ShippingLabelURL    *string
	ShippingCost        decimal.Decimal
	WeightKG            decimal.Decimal
	Dimensions          *Dimensions // JSONB data
	EstimatedDeliveryAt time.Time
	ActualDeliveryAt    *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// validate performs business rule validation.
func (f *Fulfillment) validate() error {
	if f.OrderID == uuid.Nil {
		return errors.New("order_id must not be empty")
	}

	if f.TrackingNumber == "" {
		return errors.New("tracking_number must not be empty")
	}

	if f.ShippingCost.LessThan(decimal.Zero) {
		return errors.New("shipping_cost must not be negative")
	}

	if f.WeightKG.LessThanOrEqual(decimal.Zero) {
		return errors.New("weight_kg must be greater than zero")
	}

	if f.EstimatedDeliveryAt.Before(f.CreatedAt) {
		return errors.New("estimated_delivery_at must be after created_at")
	}

	if f.CreatedAt.After(f.UpdatedAt) {
		return errors.New("created_at must be before or equal to updated_at")
	}

	// Status validation is handled by database constraints

	return nil
}

// NewFulfillment creates a new fulfillment with validation.
func NewFulfillment(
	orderID uuid.UUID,
	trackingNumber string,
	shippingCost, weightKG decimal.Decimal,
	estimatedDeliveryAt time.Time,
) (*Fulfillment, error) {
	now := time.Now()
	fulfillment := &Fulfillment{
		ID:                  uuid.New(),
		OrderID:             orderID,
		Status:              constant.FulfillmentStatusPending,
		TrackingNumber:      trackingNumber,
		ShippingCost:        shippingCost.Round(2),
		WeightKG:            weightKG.Round(2),
		EstimatedDeliveryAt: estimatedDeliveryAt,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := fulfillment.validate(); err != nil {
		return nil, err
	}

	return fulfillment, nil
}

// UpdateStatus updates the fulfillment status with validation.
func (f *Fulfillment) UpdateStatus(status constant.FulfillmentStatus) error {
	f.Status = status
	f.UpdatedAt = time.Now()

	// Set delivery timestamp for delivered status
	if status == constant.FulfillmentStatusDelivered && f.ActualDeliveryAt == nil {
		now := time.Now()
		f.ActualDeliveryAt = &now
	}

	return f.validate()
}

// SetCarrierInfo sets the carrier and shipping label information.
func (f *Fulfillment) SetCarrierInfo(carrier string, shippingLabelURL *string) error {
	f.Carrier = &carrier
	f.ShippingLabelURL = shippingLabelURL
	f.UpdatedAt = time.Now()

	return f.validate()
}

// SetDimensions sets the package dimensions.
func (f *Fulfillment) SetDimensions(dimensions *Dimensions) error {
	f.Dimensions = dimensions
	f.UpdatedAt = time.Now()

	return f.validate()
}

// SetActualDelivery sets the actual delivery time.
func (f *Fulfillment) SetActualDelivery(deliveryTime time.Time) error {
	f.ActualDeliveryAt = &deliveryTime
	f.UpdatedAt = time.Now()

	return f.validate()
}

// CanBeShipped checks if fulfillment can be shipped.
func (f *Fulfillment) CanBeShipped() bool {
	return f.Status == constant.FulfillmentStatusPending ||
		f.Status == constant.FulfillmentStatusProcessing
}

// CanBeCanceled checks if fulfillment can be canceled.
func (f *Fulfillment) CanBeCanceled() bool {
	return f.Status == constant.FulfillmentStatusPending ||
		f.Status == constant.FulfillmentStatusProcessing
}

// IsShipped checks if fulfillment is shipped.
func (f *Fulfillment) IsShipped() bool {
	return f.Status == constant.FulfillmentStatusShipped
}

// IsDelivered checks if fulfillment has been delivered.
func (f *Fulfillment) IsDelivered() bool {
	return f.Status == constant.FulfillmentStatusDelivered
}

// IsCanceled checks if fulfillment has been canceled.
func (f *Fulfillment) IsCanceled() bool {
	return f.Status == constant.FulfillmentStatusCanceled
}
