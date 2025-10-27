// Package client provides a client for interacting with the payment service.
package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"connectrpc.com/connect"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1"
	paymentv1connect "github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1/paymentv1connect"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
)

// PaymentClient defines methods available for gRPC payment operations.
type PaymentClient interface {
	CreatePaymentIntent(
		ctx context.Context,
		req *dto.CreatePaymentIntentRequest,
	) (*dto.CreatePaymentIntentResponse, error)
}

// paymentClient is a Connect-RPC client for interacting with the payment service.
type paymentClient struct {
	client paymentv1connect.PaymentServiceClient
}

// NewPaymentClient creates a new paymentClient instance with Connect-RPC.
func NewPaymentClient(
	cfg *config.Config,
) (PaymentClient, error) {
	// Create HTTP client for Connect-RPC
	httpClient := &http.Client{
		Timeout: constant.PaymentClientTimeout,
	}

	// Use static configuration for now
	baseURL := "http://" + net.JoinHostPort(
		cfg.Client.PaymentGRPCHost,
		strconv.Itoa(cfg.Client.PaymentGRPCPort),
	)

	// Create Connect-RPC client
	client := paymentv1connect.NewPaymentServiceClient(httpClient, baseURL)

	return &paymentClient{
		client: client,
	}, nil
}

// CreatePaymentIntent creates a PaymentIntent for the given order.
func (pc *paymentClient) CreatePaymentIntent(
	ctx context.Context,
	req *dto.CreatePaymentIntentRequest,
) (*dto.CreatePaymentIntentResponse, error) {
	// Build protobuf request from DTO
	pbReq := &pb.CreatePaymentIntentRequest{
		OrderId:           req.OrderID.String(),
		CustomerId:        req.CustomerID.String(),
		CustomerEmail:     req.CustomerEmail,
		Amount:            req.Amount.String(),
		Currency:          req.Currency,
		PaymentGateway:    req.PaymentGateway,
		CheckoutSessionId: req.CheckoutSessionID.String(),
		IdempotencyKey:    req.IdempotencyKey.String(),
		Items:             make([]*pb.PaymentItem, len(req.Items)),
	}

	// Convert DTO items to protobuf items
	for i, item := range req.Items {
		pbReq.Items[i] = &pb.PaymentItem{
			ProductId:   item.ProductID.String(),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			Currency:    item.Currency,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, constant.PaymentClientTimeout)
	defer cancel()

	connectReq := connect.NewRequest(pbReq)
	pkgconnect.AddAuthHeaders(ctx, connectReq)

	resp, err := pc.client.CreatePaymentIntent(ctx, connectReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call CreatePaymentIntent: %w", err)
	}

	// Convert protobuf response to DTO
	dtoResp := &dto.CreatePaymentIntentResponse{
		PaymentIntentID: resp.Msg.GetPaymentIntentId(),
		ClientSecret:    resp.Msg.GetClientSecret(),
		PaymentGateway:  resp.Msg.GetPaymentGateway(),
		Status:          resp.Msg.GetStatus(),
		Amount:          resp.Msg.GetAmount(),
		Currency:        resp.Msg.GetCurrency(),
		OrderID:         resp.Msg.GetOrderId(),
	}

	// Convert expires_at if present
	if resp.Msg.GetExpiresAt() != nil {
		expiresAt := resp.Msg.GetExpiresAt().AsTime()
		dtoResp.ExpiresAt = &expiresAt
	}

	return dtoResp, nil
}

// HealthCheck verifies the connection to payment-service.
func (pc *paymentClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, constant.PaymentClientTimeout)
	defer cancel()

	// Create a simple health check request
	req := &pb.CreatePaymentIntentRequest{
		OrderId:           "health-check-order-id",
		CustomerId:        "health-check-customer-id",
		CustomerEmail:     "health@example.com",
		Amount:            "10.00",
		Currency:          "USD",
		PaymentGateway:    "mock",
		CheckoutSessionId: "health-check-session-id",
		IdempotencyKey:    "health-check-key",
		Items:             []*pb.PaymentItem{},
	}

	connectReq := connect.NewRequest(req)
	pkgconnect.AddAuthHeaders(ctx, connectReq)

	_, err := pc.client.CreatePaymentIntent(ctx, connectReq)
	if err != nil {
		return fmt.Errorf("payment service health check failed: %w", err)
	}

	return nil
}
