// Package client provides external service clients for the order service.
package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1/fulfillmentv1connect"
	"github.com/shopspring/decimal"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
)

// FulfillmentClient defines the interface for fulfillment service integration.
type FulfillmentClient interface {
	// GetShippingCost gets the shipping cost for the order.
	GetShippingCost(
		ctx context.Context,
		order *entity.Order,
		shipping *dto.Shipping,
	) (decimal.Decimal, error)
	// WaitForFulfillmentResponse waits for fulfillment service response with timeout
	WaitForFulfillmentResponse(
		ctx context.Context,
		orderID uuid.UUID,
		timeout time.Duration,
	) (*dto.FulfillmentResponse, error)

	// NotifyWaitingSaga notifies waiting sagas about fulfillment response
	NotifyWaitingSaga(response *dto.FulfillmentResponse)
	HealthCheck(ctx context.Context) error
}

// fulfillmentClient implements FulfillmentClient using event-based correlation.
type fulfillmentClient struct {
	logger         logger.Logger
	client         fulfillmentv1connect.FulfillmentServiceClient
	fulfillmentMap map[uuid.UUID]chan *dto.FulfillmentResponse
	mutex          sync.RWMutex
}

// NewFulfillmentClient creates a new fulfillmentClient instance.
func NewFulfillmentClient(
	cfg *config.Config,
	appLogger logger.Logger,
) (FulfillmentClient, error) {
	// Create HTTP client for Connect-RPC
	httpClient := &http.Client{
		Timeout: constant.FulfillmentClientTimeout,
	}

	// Use static configuration
	baseURL := "http://" + net.JoinHostPort(
		cfg.Client.FulfillmentGRPCHost,
		strconv.Itoa(cfg.Client.FulfillmentGRPCPort),
	)

	// Create Connect-RPC client
	client := fulfillmentv1connect.NewFulfillmentServiceClient(httpClient, baseURL)

	return &fulfillmentClient{
		logger:         appLogger,
		client:         client,
		fulfillmentMap: make(map[uuid.UUID]chan *dto.FulfillmentResponse),
	}, nil
}

// GetShippingCost gets the shipping cost for the order.
func (c *fulfillmentClient) GetShippingCost(
	ctx context.Context,
	order *entity.Order,
	shipping *dto.Shipping,
) (decimal.Decimal, error) {
	c.logger.Infof("Getting shipping cost for order: %s", order.ID)

	ctx, cancel := context.WithTimeout(ctx, constant.FulfillmentClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.GetShippingCostRequest{
		Currency: order.Currency,
		Shipping: mapper.MapShippingDtoToProto(shipping),
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := c.client.GetShippingCost(ctx, req)
	if err != nil {
		c.logger.Errorf("Failed to get shipping cost for order %s: %v", order.ID, err)

		return decimal.Zero, fmt.Errorf("failed to get shipping cost: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return decimal.Zero, fmt.Errorf(
			"shipping cost calculation failed: %s",
			resp.Msg.GetErrorMessage(),
		)
	}

	shippingCost := decimal.NewFromFloat(resp.Msg.GetShippingCost())
	c.logger.Infof("Got shipping cost for order %s: %s %s", order.ID, shippingCost, order.Currency)

	return shippingCost, nil
}

// WaitForFulfillmentResponse waits for fulfillment service response with timeout.
func (c *fulfillmentClient) WaitForFulfillmentResponse(
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
	c.mutex.Lock()
	c.fulfillmentMap[orderID] = responseChan
	c.mutex.Unlock()

	// Ensure cleanup
	defer func() {
		c.mutex.Lock()
		delete(c.fulfillmentMap, orderID)
		close(responseChan)
		c.mutex.Unlock()
	}()

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		if response.Error != nil {
			c.logger.Errorf("Fulfillment failed for order %s: %v", orderID, response.Error)

			return nil, response.Error
		}

		c.logger.Infof("Received fulfillment response for order %s: ID=%s, Cost=%s, Tracking=%s",
			orderID, response.FulfillmentID, response.ShippingCost, response.TrackingNumber)

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
func (c *fulfillmentClient) NotifyWaitingSaga(response *dto.FulfillmentResponse) {
	if response == nil {
		c.logger.Warn("Received nil fulfillment response")

		return
	}

	c.mutex.RLock()
	responseChan, exists := c.fulfillmentMap[response.OrderID]
	c.mutex.RUnlock()

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

// HealthCheck verifies the connection to product-service.
func (c *fulfillmentClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, constant.FulfillmentClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.HealthRequest{})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := c.client.Health(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Msg.GetStatus() != pb.HealthStatus_HEALTH_STATUS_SERVING {
		return fmt.Errorf("service unhealthy: %s", resp.Msg.GetStatus())
	}

	return nil
}
