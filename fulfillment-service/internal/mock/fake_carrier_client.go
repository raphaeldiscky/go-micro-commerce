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
			Carrier:           constant.CarrierTypeJNE,
			Service:           "JNE Regular",
			Cost:              decimal.NewFromFloat(25000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(2 * 24 * time.Hour),
			TransitDays:       2,
		},
		{
			Carrier:           constant.CarrierTypeJT,
			Service:           "J&T Express",
			Cost:              decimal.NewFromFloat(22000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(3 * 24 * time.Hour),
			TransitDays:       3,
		},
		{
			Carrier:           constant.CarrierTypeSiCepat,
			Service:           "SiCepat REG",
			Cost:              decimal.NewFromFloat(20000),
			Currency:          "IDR",
			EstimatedDelivery: baseDate.Add(4 * 24 * time.Hour),
			TransitDays:       4,
		},
	}

	return rates, nil
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

	trackingNumber := c.generateTrackingNumber(req.Carrier)

	return &dto.ShippingLabel{
		TrackingNumber: trackingNumber,
		LabelURL:       fmt.Sprintf("https://fake-carrier.com/labels/%s.pdf", trackingNumber),
		Carrier:        req.Carrier,
		Service:        req.Service,
	}, nil
}

// GetTracking returns mock tracking information.
func (c *FakeCarrierClient) GetTracking(
	_ context.Context,
	trackingNumber string,
	carrier string,
) (*dto.TrackingInfo, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to get tracking: carrier API error")
	}

	// Simulate different tracking statuses based on tracking number
	statuses := []constant.FulfillmentStatus{
		constant.FulfillmentStatusProcessing,
		constant.FulfillmentStatusShipped,
		constant.FulfillmentStatusInTransit,
		constant.FulfillmentStatusDelivered,
	}

	// Use tracking number hash to determine status consistently
	statusIndex := len(trackingNumber) % len(statuses)
	status := statuses[statusIndex]

	location := c.generateLocation()
	description := c.generateDescription(status)

	info := &dto.TrackingInfo{
		TrackingNumber: trackingNumber,
		Status:         status,
		Carrier:        carrier,
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
	carrier string,
) error {
	time.Sleep(c.delay)

	if c.shouldFail {
		return fmt.Errorf(
			"failed to cancel shipment: TrackingNumber: %s, Carrier: %s",
			trackingNumber,
			carrier,
		)
	}

	// Simulate successful cancellation
	return nil
}

// ValidateAddress validates and normalizes an address.
func (c *FakeCarrierClient) ValidateAddress(
	_ context.Context,
	address *dto.ShippingAddress,
) (*dto.ShippingAddress, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("failed to validate address: carrier API error")
	}

	// Return the address with some mock normalization
	normalized := &dto.ShippingAddress{
		Name:       address.Name,
		Company:    address.Company,
		Address1:   fmt.Sprintf("NORMALIZED: %s", address.Address1),
		Address2:   address.Address2,
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		Phone:      address.Phone,
		Email:      address.Email,
	}

	return normalized, nil
}

// generateTrackingNumber creates a mock tracking number.
func (c *FakeCarrierClient) generateTrackingNumber(carrier string) string {
	prefix := map[string]string{
		string(constant.CarrierTypeJNE):     "JNE",
		string(constant.CarrierTypeJT):      "JT",
		string(constant.CarrierTypeSiCepat): "SC",
		string(constant.CarrierTypePOS):     "POS",
		string(constant.CarrierTypeTiki):    "TK",
	}

	carrierPrefix := prefix[carrier]
	if carrierPrefix == "" {
		carrierPrefix = "GEN"
	}

	randomSuffix := random.Int(999999999)

	return fmt.Sprintf("%s%09d", carrierPrefix, randomSuffix)
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
