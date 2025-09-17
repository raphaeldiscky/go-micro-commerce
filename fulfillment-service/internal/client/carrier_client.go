// Package client provides external service clients for the fulfillment service.
package client

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
)

// CarrierClient defines the interface for carrier service integration.
type CarrierClient interface {
	// GetRates retrieves shipping rates for a package.
	GetRates(ctx context.Context, req *dto.ShippingRequest) ([]dto.ShippingRate, error)

	// GetRate retrieves a single shipping rate for a package.
	GetRate(ctx context.Context, req *dto.ShippingRequest) (*dto.ShippingRate, error)

	// CreateShipment creates a shipment and returns a shipping label.
	CreateShipment(ctx context.Context, req *dto.ShippingRequest) (*dto.ShippingLabel, error)

	// GetTracking retrieves tracking information for a shipment.
	GetTracking(
		ctx context.Context,
		trackingNumber string,
		carrierID constant.CarrierID,
	) (*dto.TrackingInfo, error)

	// CancelShipment cancels a shipment and voids the label.
	CancelShipment(ctx context.Context, trackingNumber string, carrier constant.CarrierID) error
}
