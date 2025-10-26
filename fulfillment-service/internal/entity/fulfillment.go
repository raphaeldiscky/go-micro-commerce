// Package entity defines the Fulfillment entity and its validation logic.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// Package represents the package dimensions and weight in kg.
type Package struct {
	WeightKG decimal.Decimal `json:"weight_kg"`
	Width    decimal.Decimal `json:"width"`
	Height   decimal.Decimal `json:"height"`
	Length   decimal.Decimal `json:"length"`
	Unit     string          `json:"unit"` // cm or inch
}

// Destination represents a customer shipping address.
type Destination struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Origin represents a warehouse address.
type Origin struct {
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
}

// Fulfillment represents a fulfillment record in the marketplace.
type Fulfillment struct {
	ID                  uuid.UUID
	OrderID             uuid.UUID // Reference to order from order-service
	Status              constant.FulfillmentStatus
	TrackingNumber      string
	CourierID           constant.CourierID
	ShippingLabelURL    string
	Currency            string
	ShippingCost        decimal.Decimal
	Destination         Destination
	Origin              Origin
	Package             Package
	EstimatedDeliveryAt time.Time
	ActualDeliveryAt    *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewFulfillment creates a new fulfillment with validation.
func NewFulfillment(
	orderID uuid.UUID,
	trackingNumber, currency string,
	shippingCost decimal.Decimal,
	packageData Package,
	destination Destination,
	origin Origin,
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
		Package:             packageData,
		Destination:         destination,
		Origin:              origin,
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

// SetCourierInfo sets the Courier and shipping label information.
func (f *Fulfillment) SetCourierInfo(courierID constant.CourierID, shippingLabelURL string) error {
	f.CourierID = courierID
	f.ShippingLabelURL = shippingLabelURL
	f.UpdatedAt = time.Now()

	return f.validate()
}

// SetPackage sets the package dimensions and weight.
func (f *Fulfillment) SetPackage(packageData Package) error {
	f.Package = packageData
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

	if f.EstimatedDeliveryAt.Before(f.CreatedAt) {
		return errors.New("estimated_delivery_at must be after created_at")
	}

	if f.CreatedAt.After(f.UpdatedAt) {
		return errors.New("created_at must be before or equal to updated_at")
	}

	// Status validation is handled by database constraints

	return nil
}
