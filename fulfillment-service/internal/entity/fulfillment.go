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
	Width  decimal.Decimal `json:"width"`
	Height decimal.Decimal `json:"height"`
	Length decimal.Decimal `json:"length"`
	Unit   string          `json:"unit"` // cm or inch
}

// ToAddress represents a customer shipping address.
type ToAddress struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// FromAddress represents a warehouse address.
type FromAddress struct {
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Fulfillment represents a fulfillment record in the marketplace.
type Fulfillment struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID // Reference to order from order-service
	Status              constant.FulfillmentStatus
	TrackingNumber      string
	CarrierID           constant.CarrierID
	ShippingLabelURL    string
	Currency            string
	ShippingCost        decimal.Decimal
	ToAddress           ToAddress
	FromAddress         FromAddress
	WeightKG            decimal.Decimal
	Dimensions          Dimensions // JSONB data
	EstimatedDeliveryAt time.Time
	ActualDeliveryAt    *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewFulfillment creates a new fulfillment with validation.
func NewFulfillment(
	orderID uuid.UUID,
	trackingNumber, currency string,
	shippingCost, weightKG decimal.Decimal,
	fromAddress FromAddress,
	toAddress ToAddress,
	estimatedDeliveryAt time.Time,
) (*Fulfillment, error) {
	now := time.Now()
	fulfillment := &Fulfillment{
		ID:                  uuid.New(),
		OrderID:             orderID,
		Status:              constant.FulfillmentStatusPending,
		TrackingNumber:      trackingNumber,
		Currency:            currency,
		ShippingCost:        shippingCost.Round(constant.DefaultPricingScale),
		WeightKG:            weightKG.Round(constant.DefaultPricingScale),
		EstimatedDeliveryAt: estimatedDeliveryAt,
		ToAddress:           toAddress,
		FromAddress:         fromAddress,
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
func (f *Fulfillment) SetCarrierInfo(carrierID constant.CarrierID, shippingLabelURL string) error {
	f.CarrierID = carrierID
	f.ShippingLabelURL = shippingLabelURL
	f.UpdatedAt = time.Now()

	return f.validate()
}

// SetDimensions sets the package dimensions.
func (f *Fulfillment) SetDimensions(dimensions Dimensions) error {
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
