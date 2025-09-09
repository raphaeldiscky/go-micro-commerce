// Package mock provides a mock implementation of the CarrierClient interface.
package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/random"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
)

// FakeCarrierClient provides a mock implementation of CarrierClientInterface for testing.
type FakeCarrierClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakeCarrierClient creates a new instance of FakeCarrierClient.
func NewFakeCarrierClient() *FakeCarrierClient {
	return &FakeCarrierClient{
		shouldFail: false,
		delay:      time.Millisecond * 100, // Simulate network delay
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *FakeCarrierClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// SetDelay configures the simulated network delay.
func (c *FakeCarrierClient) SetDelay(delay time.Duration) {
	c.delay = delay
}

// GetRates returns mock shipping rates for different carriers.
func (c *FakeCarrierClient) GetRates(
	_ context.Context,
	_ *dto.ShippingRequest,
) ([]dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("simulated carrier API error")
	}

	baseDate := time.Now().Add(24 * time.Hour)

	rates := []dto.ShippingRate{
		{
			CarrierID:         constant.CarrierJNE,
			Service:           "JNE Regular",
			ShippingCost:      decimal.NewFromFloat(25000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(2 * 24 * time.Hour),
			TransitDays:       2,
		},
		{
			CarrierID:         constant.CarrierJT,
			Service:           "J&T Express",
			ShippingCost:      decimal.NewFromFloat(22000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(3 * 24 * time.Hour),
			TransitDays:       3,
		},
		{
			CarrierID:         constant.CarrierSiCepat,
			Service:           "SiCepat REG",
			ShippingCost:      decimal.NewFromFloat(20000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(4 * 24 * time.Hour),
			TransitDays:       3,
		},
	}

	return rates, nil
}

// GetRate returns a mock shipping rate for a specific carrier.
func (c *FakeCarrierClient) GetRate(
	ctx context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("simulated carrier API error")
	}

	rates, err := c.GetRates(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, rate := range rates {
		if rate.CarrierID == req.CarrierID {
			return &rate, nil
		}
	}

	return nil, fmt.Errorf("shipping rate not found")
}

// CreateShipment creates a mock shipping label.
func (c *FakeCarrierClient) CreateShipment(
	_ context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingLabel, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to create shipment: carrier API error")
	}

	carrierService := c.getCarrierInfo(req.CarrierID)
	trackingNumber := c.generateTrackingNumber(req.CarrierID)

	return &dto.ShippingLabel{
		TrackingNumber: trackingNumber,
		LabelURL:       fmt.Sprintf("https://fake-carrier.com/labels/%s.pdf", trackingNumber),
		CarrierID:      req.CarrierID,
		Service:        carrierService,
	}, nil
}

// getCarrierInfo get carrier name and service.
func (c *FakeCarrierClient) getCarrierInfo(carrierID constant.CarrierID) string {
	switch carrierID {
	case constant.CarrierJNE:
		return "JNE Regular"
	case constant.CarrierJT:
		return "J&T Express"
	case constant.CarrierSiCepat:
		return "SiCepat REG"
	case constant.CarrierPOS:
		return "POS Indonesia"
	default:
		return ""
	}
}

// GetTracking returns mock tracking information.
func (c *FakeCarrierClient) GetTracking(
	_ context.Context,
	trackingNumber string,
	carrierID constant.CarrierID,
) (*dto.TrackingInfo, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to get tracking: carrier API error")
	}

	statuses := []constant.FulfillmentStatus{
		constant.FulfillmentStatusProcessing,
		constant.FulfillmentStatusShipped,
		constant.FulfillmentStatusInTransit,
		constant.FulfillmentStatusDelivered,
	}

	statusIndex := len(trackingNumber) % len(statuses)
	status := statuses[statusIndex]

	location := c.generateLocation()
	description := c.generateDescription(status)

	info := &dto.TrackingInfo{
		TrackingNumber: trackingNumber,
		Status:         status,
		CarrierID:      carrierID,
		LastUpdate:     time.Now().Add(-time.Duration(random.Int(24)) * time.Hour),
		Location:       location,
		Description:    description,
	}

	if status == constant.FulfillmentStatusDelivered {
		deliveredAt := time.Now().Add(-time.Duration(random.Int(48)) * time.Hour)
		info.DeliveredAt = &deliveredAt
	}

	return info, nil
}

// CancelShipment cancels a mock shipment.
func (c *FakeCarrierClient) CancelShipment(
	_ context.Context,
	trackingNumber string,
	carrierID constant.CarrierID,
) error {
	time.Sleep(c.delay)

	if c.shouldFail {
		return fmt.Errorf(
			"failed to cancel shipment: TrackingNumber: %s, CarrierID: %s",
			trackingNumber,
			carrierID,
		)
	}

	return nil
}

// generateTrackingNumber creates a mock tracking number.
func (c *FakeCarrierClient) generateTrackingNumber(carrierID constant.CarrierID) string {
	prefix := carrierID
	randomSuffix := random.Int(999999999)

	return fmt.Sprintf("%s-%09d", prefix, randomSuffix)
}

// generateLocation creates a mock location.
func (c *FakeCarrierClient) generateLocation() string {
	locations := []string{
		"Jakarta, Indonesia",
		"Surabaya, Indonesia",
		"Bandung, Indonesia",
		"Medan, Indonesia",
		"Semarang, Indonesia",
		"Distribution Center",
		"Local Facility",
	}

	return locations[random.Int(int64(len(locations)))]
}

// generateDescription creates a status description.
func (c *FakeCarrierClient) generateDescription(status constant.FulfillmentStatus) string {
	descriptions := map[constant.FulfillmentStatus]string{
		constant.FulfillmentStatusProcessing: "Package is being prepared for shipment",
		constant.FulfillmentStatusShipped:    "Package has been picked up by carrier",
		constant.FulfillmentStatusInTransit:  "Package is in transit to destination",
		constant.FulfillmentStatusDelivered:  "Package has been delivered successfully",
	}

	if desc, exists := descriptions[status]; exists {
		return desc
	}

	return "Status update available"
}
