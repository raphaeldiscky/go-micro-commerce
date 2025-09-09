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

// PaymentClientInterface defines the interface for payment service integration.
type PaymentClientInterface interface {
	// WaitForPaymentResponse waits for payment service response with timeout
	WaitForPaymentResponse(
		ctx context.Context,
		orderID uuid.UUID,
		timeout time.Duration,
	) (*dto.PaymentResponse, error)

	// NotifyWaitingSaga notifies waiting sagas about payment response
	NotifyWaitingSaga(response *dto.PaymentResponse)

	// Close cleans up resources
	Close() error
}

// PaymentClient implements PaymentClientInterface using event-based correlation.
type PaymentClient struct {
	logger     logger.Logger
	paymentMap map[uuid.UUID]chan *dto.PaymentResponse
	mutex      sync.RWMutex
}

// NewPaymentClient creates a new PaymentClient instance.
func NewPaymentClient(appLogger logger.Logger) PaymentClientInterface {
	return &PaymentClient{
		logger:     appLogger,
		paymentMap: make(map[uuid.UUID]chan *dto.PaymentResponse),
	}
}

// WaitForPaymentResponse waits for payment service response with timeout.
func (c *PaymentClient) WaitForPaymentResponse(
	ctx context.Context,
	orderID uuid.UUID,
	timeout time.Duration,
) (*dto.PaymentResponse, error) {
	c.logger.Infof(
		"Waiting for payment response for order: %s with timeout: %v",
		orderID,
		timeout,
	)

	// Create response channel for this order
	responseChan := make(chan *dto.PaymentResponse, 1)

	// Register correlation
	c.mutex.Lock()
	c.paymentMap[orderID] = responseChan
	c.mutex.Unlock()

	// Ensure cleanup
	defer func() {
		c.mutex.Lock()
		delete(c.paymentMap, orderID)
		close(responseChan)
		c.mutex.Unlock()
	}()

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		if response.Error != nil {
			c.logger.Errorf("Payment failed for order %s: %v", orderID, response.Error)

			return nil, response.Error
		}

		c.logger.Infof("Received payment response for order %s: ID=%s, Status=%s",
			orderID, response.PaymentID, response.Status)

		return response, nil

	case <-time.After(timeout):
		c.logger.Warnf("Timeout waiting for payment response for order: %s", orderID)

		return nil, fmt.Errorf(
			"timeout waiting for payment response for order %s after %v",
			orderID,
			timeout,
		)

	case <-ctx.Done():
		c.logger.Warnf(
			"Context canceled while waiting for payment response for order: %s",
			orderID,
		)

		return nil, ctx.Err()
	}
}

// NotifyWaitingSaga notifies waiting sagas about payment response.
func (c *PaymentClient) NotifyWaitingSaga(response *dto.PaymentResponse) {
	if response == nil {
		c.logger.Warn("Received nil payment response")

		return
	}

	c.mutex.RLock()
	responseChan, exists := c.paymentMap[response.OrderID]
	c.mutex.RUnlock()

	if !exists {
		c.logger.Debugf(
			"No waiting saga found for payment response of order: %s",
			response.OrderID,
		)

		return
	}

	// Non-blocking send to avoid deadlock
	select {
	case responseChan <- response:
		c.logger.Infof(
			"Notified waiting saga for order %s about payment response",
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
func (c *PaymentClient) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Close all pending channels
	for orderID, ch := range c.paymentMap {
		c.logger.Infof("Closing pending payment correlation for order: %s", orderID)
		close(ch)
	}

	// Clear the map
	c.paymentMap = make(map[uuid.UUID]chan *dto.PaymentResponse)

	return nil
}
