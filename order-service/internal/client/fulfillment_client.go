package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// FulfillmentClientInterface defines the interface for fulfillment service integration.
type FulfillmentClientInterface interface {
	// WaitForFulfillmentResponse waits for fulfillment service response with timeout
	WaitForFulfillmentResponse(
		ctx context.Context,
		orderID uuid.UUID,
		timeout time.Duration,
	) (*dto.FulfillmentResponse, error)

	// NotifyWaitingSaga notifies waiting sagas about fulfillment response
	NotifyWaitingSaga(response *dto.FulfillmentResponse)

	// Close cleans up resources
	Close() error
}

// FulfillmentClient implements FulfillmentClientInterface using event-based correlation.
type FulfillmentClient struct {
	logger           logger.Logger
	correlationMap   map[uuid.UUID]chan *dto.FulfillmentResponse
	correlationMutex sync.RWMutex
}

// NewFulfillmentClient creates a new FulfillmentClient instance.
func NewFulfillmentClient(appLogger logger.Logger) FulfillmentClientInterface {
	return &FulfillmentClient{
		logger:         appLogger,
		correlationMap: make(map[uuid.UUID]chan *dto.FulfillmentResponse),
	}
}

// WaitForFulfillmentResponse waits for fulfillment service response with timeout.
func (c *FulfillmentClient) WaitForFulfillmentResponse(
	ctx context.Context,
	orderID uuid.UUID,
	timeout time.Duration,
) (*dto.FulfillmentResponse, error) {
	c.logger.Infof(
		"Waiting for fulfillment response for order: %s with timeout: %v",
		orderID,
		timeout,
	)

	// Create response channel for this order
	responseChan := make(chan *dto.FulfillmentResponse, 1)

	// Register correlation
	c.correlationMutex.Lock()
	c.correlationMap[orderID] = responseChan
	c.correlationMutex.Unlock()

	// Ensure cleanup
	defer func() {
		c.correlationMutex.Lock()
		delete(c.correlationMap, orderID)
		close(responseChan)
		c.correlationMutex.Unlock()
	}()

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		if response.Error != nil {
			c.logger.Errorf("Fulfillment failed for order %s: %v", orderID, response.Error)

			return nil, response.Error
		}

		c.logger.Infof("Received fulfillment response for order %s: ID=%s, Tracking=%s",
			orderID, response.FulfillmentID, response.TrackingNumber)

		return response, nil

	case <-time.After(timeout):
		c.logger.Warnf("Timeout waiting for fulfillment response for order: %s", orderID)

		return nil, fmt.Errorf(
			"timeout waiting for fulfillment response for order %s after %v",
			orderID,
			timeout,
		)

	case <-ctx.Done():
		c.logger.Warnf(
			"Context canceled while waiting for fulfillment response for order: %s",
			orderID,
		)

		return nil, ctx.Err()
	}
}

// NotifyWaitingSaga notifies waiting sagas about fulfillment response.
func (c *FulfillmentClient) NotifyWaitingSaga(response *dto.FulfillmentResponse) {
	if response == nil {
		c.logger.Warn("Received nil fulfillment response")

		return
	}

	c.correlationMutex.RLock()
	responseChan, exists := c.correlationMap[response.OrderID]
	c.correlationMutex.RUnlock()

	if !exists {
		c.logger.Debugf(
			"No waiting saga found for fulfillment response of order: %s",
			response.OrderID,
		)

		return
	}

	// Non-blocking send to avoid deadlock
	select {
	case responseChan <- response:
		c.logger.Infof(
			"Notified waiting saga for order %s about fulfillment response",
			response.OrderID,
		)
	default:
		c.logger.Warnf(
			"Failed to notify saga for order %s - channel full or closed",
			response.OrderID,
		)
	}
}

// Close cleans up resources.
func (c *FulfillmentClient) Close() error {
	c.correlationMutex.Lock()
	defer c.correlationMutex.Unlock()

	// Close all pending channels
	for orderID, ch := range c.correlationMap {
		c.logger.Infof("Closing pending fulfillment correlation for order: %s", orderID)
		close(ch)
	}

	// Clear the map
	c.correlationMap = make(map[uuid.UUID]chan *dto.FulfillmentResponse)

	return nil
}
