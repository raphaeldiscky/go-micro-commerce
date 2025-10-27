// Package client provides a client for interacting with the fulfillment service.
package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"connectrpc.com/connect"
	"github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1/fulfillmentv1connect"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/fulfillment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
)

// FulfillmentClient defines methods available for gRPC fulfillment operations.
type FulfillmentClient interface {
	GetShippingCost(
		ctx context.Context,
		req *dto.GetShippingCostRequest,
	) (*dto.GetShippingCostResponse, error)
	HealthCheck(ctx context.Context) error
}

// fulfillmentClient is a Connect-RPC client for interacting with the fulfillment service.
type fulfillmentClient struct {
	client fulfillmentv1connect.FulfillmentServiceClient
}

// NewFulfillmentClient creates a new fulfillmentClient instance with Connect-RPC.
func NewFulfillmentClient(
	cfg *config.Config,
) (FulfillmentClient, error) {
	// Create HTTP client for Connect-RPC
	httpClient := &http.Client{
		Timeout: constant.FulfillmentClientTimeout,
	}

	// Use static configuration for now
	baseURL := "http://" + net.JoinHostPort(
		cfg.Client.FulfillmentGRPCHost,
		strconv.Itoa(cfg.Client.FulfillmentGRPCPort),
	)

	// Create Connect-RPC client
	client := fulfillmentv1connect.NewFulfillmentServiceClient(httpClient, baseURL)

	return &fulfillmentClient{
		client: client,
	}, nil
}

// GetShippingCost calculates shipping cost for the given parameters.
func (c *fulfillmentClient) GetShippingCost(
	ctx context.Context,
	req *dto.GetShippingCostRequest,
) (*dto.GetShippingCostResponse, error) {
	// Create the protobuf request
	pbReq := &pb.GetShippingCostRequest{
		Currency: req.Currency,
		Courier: &pb.Courier{
			CourierId: req.CourierID,
		},
		Destination: &pb.Destination{
			City:        req.DestinationCity,
			State:       req.DestinationState,
			PostalCode:  req.DestinationPostalCode,
			CountryCode: req.DestinationCountryCode,
		},
		Origin: &pb.Origin{
			City:        req.OriginCity,
			State:       req.OriginState,
			PostalCode:  req.OriginPostalCode,
			CountryCode: req.OriginCountryCode,
		},
		Package: &pb.Package{
			WeightKg: req.WeightKG,
			Width:    req.Width,
			Height:   req.Height,
			Length:   req.Length,
			Unit:     req.Unit,
		},
	}

	ctx, cancel := context.WithTimeout(ctx, constant.FulfillmentClientTimeout)
	defer cancel()

	connectReq := connect.NewRequest(pbReq)
	pkgconnect.AddAuthHeaders(ctx, connectReq)

	pbResp, err := c.client.GetShippingCost(ctx, connectReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetShippingCost: %w", err)
	}

	resp := pbResp.Msg

	// Map the protobuf response to our DTO
	if !resp.GetSuccess() {
		return &dto.GetShippingCostResponse{
			Success:      false,
			ErrorMessage: resp.GetErrorMessage(),
		}, nil
	}

	return &dto.GetShippingCostResponse{
		Success:      true,
		ShippingCost: resp.GetShippingCost(),
		Currency:     resp.GetCurrency(),
	}, nil
}

// HealthCheck verifies the connection to fulfillment-service.
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
