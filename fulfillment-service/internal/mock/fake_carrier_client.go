// Package mock provides a mock implementation of the CourierClient interface.
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
	fakeCourierDelay             = time.Millisecond * 100
	fakeCourierBaseDateAdd       = 24 * time.Hour
	fakeCourierShippingCost      = 25000
	fakeCourierTransitDays       = 2
	fakeCourierLastUpdateAdd     = 24 * time.Hour
	fakeCourierDeliveredAtAdd    = 24 * time.Hour
	fakeCourierTrackingNumberMax = 999999999
)

// fakeCourierClient provides a mock implementation of CourierClient interface for testing.
type fakeCourierClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakeCourierClient creates a new instance of fakeCourierClient.
func NewFakeCourierClient() client.CourierClient {
	return &fakeCourierClient{
		shouldFail: false,
		delay:      fakeCourierDelay, // Simulate network delay
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *fakeCourierClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// SetDelay configures the simulated network delay.
func (c *fakeCourierClient) SetDelay(delay time.Duration) {
	c.delay = delay
}

// GetRates returns mock shipping rates for different Couriers.
func (c *fakeCourierClient) GetRates(
	_ context.Context,
	_ *dto.ShippingRequest,
) ([]dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("simulated Courier API error")
	}

	baseDate := time.Now().Add(fakeCourierBaseDateAdd)

	rates := []dto.ShippingRate{
		{
			CourierID:          constant.CourierJNE,
			CourierServiceName: "JNE Regular",
			ShippingCost:       decimal.NewFromFloat(fakeCourierShippingCost),
			Currency:           "USD",
			EstimatedDelivery:  baseDate.Add(2 * 24 * time.Hour),
			TransitDays:        fakeCourierTransitDays,
		},
		{
			CourierID:          constant.CourierJT,
			CourierServiceName: "J&T Express",
			ShippingCost:       decimal.NewFromFloat(fakeCourierShippingCost),
			Currency:           "USD",
			EstimatedDelivery:  baseDate.Add(3 * 24 * time.Hour),
			TransitDays:        fakeCourierTransitDays,
		},
		{
			CourierID:          constant.CourierSiCepat,
			CourierServiceName: "SiCepat REG",
			ShippingCost:       decimal.NewFromFloat(fakeCourierShippingCost),
			Currency:           "USD",
			EstimatedDelivery:  baseDate.Add(4 * 24 * time.Hour),
			TransitDays:        fakeCourierTransitDays,
		},
	}

	return rates, nil
}

// GetRate returns a mock shipping rate for a specific Courier.
func (c *fakeCourierClient) GetRate(
	ctx context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingRate, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("simulated Courier API error")
	}

	rates, err := c.GetRates(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, rate := range rates {
		if rate.CourierID == req.CourierID {
			return &rate, nil
		}
	}

	return nil, errors.New("shipping rate not found")
}

// CreateShipment creates a mock shipping label.
func (c *fakeCourierClient) CreateShipment(
	_ context.Context,
	req *dto.ShippingRequest,
) (*dto.ShippingLabel, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to create shipment: Courier API error")
	}

	courierServiceName := c.getCourierInfo(req.CourierID)
	trackingNumber := c.generateTrackingNumber(req.CourierID)

	return &dto.ShippingLabel{
		TrackingNumber:     trackingNumber,
		LabelURL:           fmt.Sprintf("https://fake-Courier.com/labels/%s.pdf", trackingNumber),
		CourierID:          req.CourierID,
		CourierServiceName: courierServiceName,
	}, nil
}

// getCourierInfo get Courier name and service.
func (c *fakeCourierClient) getCourierInfo(courierID constant.CourierID) string {
	switch courierID {
	case constant.CourierJNE:
		return "JNE Regular"
	case constant.CourierJT:
		return "J&T Express"
	case constant.CourierSiCepat:
		return "SiCepat REG"
	case constant.CourierPOS:
		return "POS Indonesia"
	default:
		return ""
	}
}

// GetTracking returns mock tracking information.
func (c *fakeCourierClient) GetTracking(
	_ context.Context,
	trackingNumber string,
	courierID constant.CourierID,
) (*dto.TrackingInfo, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("failed to get tracking: Courier API error")
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
		CourierID:      courierID,
		LastUpdate:     time.Now().Add(fakeCourierLastUpdateAdd),
		Location:       location,
		Description:    description,
	}

	if status == constant.FulfillmentStatusDelivered {
		deliveredAt := time.Now().Add(fakeCourierDeliveredAtAdd)
		info.DeliveredAt = &deliveredAt
	}

	return info, nil
}

// CancelShipment cancels a mock shipment.
func (c *fakeCourierClient) CancelShipment(
	_ context.Context,
	trackingNumber string,
	courierID constant.CourierID,
) error {
	time.Sleep(c.delay)

	if c.shouldFail {
		return fmt.Errorf(
			"failed to cancel shipment: TrackingNumber: %s, CourierID: %s",
			trackingNumber,
			courierID,
		)
	}

	return nil
}

// generateTrackingNumber creates a mock tracking number.
func (c *fakeCourierClient) generateTrackingNumber(courierID constant.CourierID) string {
	prefix := courierID
	randomSuffix := random.Int(fakeCourierTrackingNumberMax)

	return fmt.Sprintf("%s-%09d", prefix, randomSuffix)
}

// generateLocation creates a mock location.
func (c *fakeCourierClient) generateLocation() string {
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
func (c *fakeCourierClient) generateDescription(status constant.FulfillmentStatus) string {
	descriptions := map[constant.FulfillmentStatus]string{
		constant.FulfillmentStatusProcessing: "Package is being prepared for shipment",
		constant.FulfillmentStatusShipped:    "Package has been picked up by Courier",
		constant.FulfillmentStatusInTransit:  "Package is in transit to destination",
		constant.FulfillmentStatusDelivered:  "Package has been delivered successfully",
	}

	if desc, exists := descriptions[status]; exists {
		return desc
	}

	return "Status update available"
}
