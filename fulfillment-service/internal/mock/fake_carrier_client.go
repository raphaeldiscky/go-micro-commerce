// Package mock provides a mock implementation of the CarrierClient interface.
package mock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/random"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/dto"
)

const (
	fakeCarrierDelay             = time.Millisecond * 100
	fakeCarrierBaseDateAdd       = 24 * time.Hour
	fakeCarrierShippingCost      = 25000
	fakeCarrierTransitDays       = 2
	fakeCarrierLastUpdateAdd     = 24 * time.Hour
	fakeCarrierDeliveredAtAdd    = 24 * time.Hour
	fakeCarrierTrackingNumberMax = 999999999
)

// fakeCarrierClient provides a mock implementation of CarrierClient interface for testing.
type fakeCarrierClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakeCarrierClient creates a new instance of fakeCarrierClient.
func NewFakeCarrierClient() client.CarrierClient {
	return &fakeCarrierClient{
		shouldFail: false,
		delay:      fakeCarrierDelay, // Simulate network delay
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *fakeCarrierClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// SetDelay configures the simulated network delay.
func (c *fakeCarrierClient) SetDelay(delay time.Duration) {
	c.delay = delay
}

// GetRates returns mock shipping rates for different carriers.
func (c *fakeCarrierClient) GetRates(
	_ context.Context,
	_ *dto.ShippingRequest,
) ([]dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("simulated carrier API error")
	}

	baseDate := time.Now().Add(fakeCarrierBaseDateAdd)

	rates := []dto.ShippingRate{
		{
			CarrierID:         constant.CarrierJNE,
			Service:           "JNE Regular",
			ShippingCost:      decimal.NewFromFloat(fakeCarrierShippingCost),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(2 * 24 * time.Hour),
			TransitDays:       fakeCarrierTransitDays,
		},
		{
			CarrierID:         constant.CarrierJT,
			Service:           "J&T Express",
			ShippingCost:      decimal.NewFromFloat(fakeCarrierShippingCost),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(3 * 24 * time.Hour),
			TransitDays:       fakeCarrierTransitDays,
		},
		{
			CarrierID:         constant.CarrierSiCepat,
			Service:           "SiCepat REG",
			ShippingCost:      decimal.NewFromFloat(fakeCarrierShippingCost),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(4 * 24 * time.Hour),
			TransitDays:       fakeCarrierTransitDays,
		},
	}

	return rates, nil
}

// GetRate returns a mock shipping rate for a specific carrier.
func (c *fakeCarrierClient) GetRate(
	ctx context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("simulated carrier API error")
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

	return nil, errors.New("shipping rate not found")
}

// CreateShipment creates a mock shipping label.
func (c *fakeCarrierClient) CreateShipment(
	_ context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingLabel, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to create shipment: carrier API error")
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
func (c *fakeCarrierClient) getCarrierInfo(carrierID constant.CarrierID) string {
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
func (c *fakeCarrierClient) GetTracking(
	_ context.Context,
	trackingNumber string,
	carrierID constant.CarrierID,
) (*dto.TrackingInfo, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to get tracking: carrier API error")
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
		LastUpdate:     time.Now().Add(fakeCarrierLastUpdateAdd),
		Location:       location,
		Description:    description,
	}

	if status == constant.FulfillmentStatusDelivered {
		deliveredAt := time.Now().Add(fakeCarrierDeliveredAtAdd)
		info.DeliveredAt = &deliveredAt
	}

	return info, nil
}

// CancelShipment cancels a mock shipment.
func (c *fakeCarrierClient) CancelShipment(
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
func (c *fakeCarrierClient) generateTrackingNumber(carrierID constant.CarrierID) string {
	prefix := carrierID
	randomSuffix := random.Int(fakeCarrierTrackingNumberMax)

	return fmt.Sprintf("%s-%09d", prefix, randomSuffix)
}

// generateLocation creates a mock location.
func (c *fakeCarrierClient) generateLocation() string {
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
func (c *fakeCarrierClient) generateDescription(status constant.FulfillmentStatus) string {
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
