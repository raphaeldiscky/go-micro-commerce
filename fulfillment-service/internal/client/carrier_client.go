// Package client provides external service clients for the fulfillment service.
package client

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// ShippingRate represents the cost and estimated delivery for a shipping option.
type ShippingRate struct {
	Carrier           constant.CarrierType `json:"carrier"`
	Service           string               `json:"service"`
	Cost              decimal.Decimal      `json:"cost"`
	Currency          string               `json:"currency"`
	EstimatedDelivery time.Time            `json:"estimated_delivery"`
	TransitDays       int                  `json:"transit_days"`
}

// ShippingLabel represents a shipping label created by a carrier.
type ShippingLabel struct {
	TrackingNumber string `json:"tracking_number"`
	LabelURL       string `json:"label_url"`
	Carrier        string `json:"carrier"`
	Service        string `json:"service"`
}

// TrackingInfo represents the current status of a shipment.
type TrackingInfo struct {
	TrackingNumber string                     `json:"tracking_number"`
	Status         constant.FulfillmentStatus `json:"status"`
	Carrier        string                     `json:"carrier"`
	LastUpdate     time.Time                  `json:"last_update"`
	Location       string                     `json:"location,omitempty"`
	Description    string                     `json:"description,omitempty"`
	DeliveredAt    *time.Time                 `json:"delivered_at,omitempty"`
}

// ShippingAddress represents an address for shipping.
type ShippingAddress struct {
	Name       string `json:"name"`
	Company    string `json:"company,omitempty"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
}

// Package represents the dimensions and weight of a package.
type Package struct {
	Weight     decimal.Decimal            `json:"weight"`     // in kg
	Dimensions map[string]decimal.Decimal `json:"dimensions"` // width, height, length in cm
}

// ShippingRequest represents a request to create a shipping label.
type ShippingRequest struct {
	OrderID         uuid.UUID       `json:"order_id"`
	Carrier         string          `json:"carrier"`
	Service         string          `json:"service"`
	FromAddress     ShippingAddress `json:"from_address"`
	ToAddress       ShippingAddress `json:"to_address"`
	Package         Package         `json:"package"`
	InsuranceAmount decimal.Decimal `json:"insurance_amount,omitempty"`
	Signature       bool            `json:"signature,omitempty"`
}

// CarrierClientInterface defines the interface for carrier service integration.
type CarrierClientInterface interface {
	// GetRates retrieves shipping rates for a package.
	GetRates(ctx context.Context, req *ShippingRequest) ([]ShippingRate, error)

	// CreateShipment creates a shipment and returns a shipping label.
	CreateShipment(ctx context.Context, req *ShippingRequest) (*ShippingLabel, error)

	// GetTracking retrieves tracking information for a shipment.
	GetTracking(ctx context.Context, trackingNumber string, carrier string) (*TrackingInfo, error)

	// CancelShipment cancels a shipment and voids the label.
	CancelShipment(ctx context.Context, trackingNumber string, carrier string) error

	// ValidateAddress validates a shipping address.
	ValidateAddress(ctx context.Context, address *ShippingAddress) (*ShippingAddress, error)
}
