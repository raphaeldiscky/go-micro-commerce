package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1/paymentv1connect"
	"github.com/shopspring/decimal"

	pkgconnect "github.com/raphaeldiscky/go-micro-commerce/pkg/connect"
	pb "github.com/raphaeldiscky/go-micro-commerce/proto/payment/v1"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PaymentIntentResponse represents the response from creating a payment intent.
type PaymentIntentResponse struct {
	PaymentID            uuid.UUID
	OrderID              uuid.UUID
	Amount               decimal.Decimal
	Currency             string
	PaymentGateway       string
	GatewayTransactionID string
	GatewayMetadata      map[string]any
	ExpiresAt            *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// PaymentClientGRPC defines methods for synchronous payment operations via gRPC.
type PaymentClientGRPC interface {
	CreatePaymentIntent(
		ctx context.Context,
		orderID uuid.UUID,
		amount decimal.Decimal,
		currency string,
		paymentGateway string,
		customerID uuid.UUID,
		customerEmail string,
	) (*PaymentIntentResponse, error)
	HealthCheck(ctx context.Context) error
}

// paymentClientGRPC is a Connect-RPC client for synchronous payment service interaction.
type paymentClientGRPC struct {
	client paymentv1connect.PaymentServiceClient
}

// NewPaymentClientGRPC creates a new paymentClientGRPC instance with Connect-RPC.
func NewPaymentClientGRPC(
	cfg *config.Config,
) (PaymentClientGRPC, error) {
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

	return &paymentClientGRPC{
		client: client,
	}, nil
}

// CreatePaymentIntent creates a payment intent synchronously.
func (pc *paymentClientGRPC) CreatePaymentIntent(
	ctx context.Context,
	orderID uuid.UUID,
	amount decimal.Decimal,
	currency string,
	paymentGateway string,
	customerID uuid.UUID,
	customerEmail string,
) (*PaymentIntentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, constant.PaymentClientTimeout)
	defer cancel()

	// Map payment gateway string to proto enum
	var paymentGatewayEnum pb.PaymentGateway

	switch paymentGateway {
	case string(constant.PaymentGatewayStripe):
		paymentGatewayEnum = pb.PaymentGateway_PAYMENT_GATEWAY_STRIPE
	case string(constant.PaymentGatewayXendit):
		paymentGatewayEnum = pb.PaymentGateway_PAYMENT_GATEWAY_XENDIT
	default:
		paymentGatewayEnum = pb.PaymentGateway_PAYMENT_GATEWAY_STRIPE
	}

	req := connect.NewRequest(&pb.CreatePaymentIntentRequest{
		OrderId:        orderID.String(),
		Amount:         amount.InexactFloat64(),
		Currency:       currency,
		PaymentGateway: paymentGatewayEnum,
		CustomerEmail:  customerEmail,
		CustomerId:     customerID.String(),
	})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.CreatePaymentIntent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call CreatePaymentIntent: %w", err)
	}

	// Parse payment ID
	paymentID, err := uuid.Parse(resp.Msg.GetPaymentId())
	if err != nil {
		return nil, fmt.Errorf("invalid payment_id: %w", err)
	}

	// Parse order ID
	respOrderID, err := uuid.Parse(resp.Msg.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("invalid order_id: %w", err)
	}

	// Convert gateway_metadata from proto Struct to map[string]any
	gatewayMetadata := make(map[string]any)
	if resp.Msg.GetGatewayMetadata() != nil {
		gatewayMetadata = resp.Msg.GetGatewayMetadata().AsMap()
	}

	// Parse expires_at if present
	var expiresAt *time.Time

	if resp.Msg.GetExpiresAt() != nil {
		expiry := resp.Msg.GetExpiresAt().AsTime()
		expiresAt = &expiry
	}

	return &PaymentIntentResponse{
		PaymentID:            paymentID,
		OrderID:              respOrderID,
		Amount:               decimal.NewFromFloat(resp.Msg.GetAmount()),
		Currency:             resp.Msg.GetCurrency(),
		PaymentGateway:       resp.Msg.GetPaymentGateway().String(),
		GatewayTransactionID: resp.Msg.GetGatewayTransactionId(),
		GatewayMetadata:      gatewayMetadata,
		ExpiresAt:            expiresAt,
		CreatedAt:            resp.Msg.GetCreatedAt().AsTime(),
		UpdatedAt:            resp.Msg.GetUpdatedAt().AsTime(),
	}, nil
}

// HealthCheck verifies the connection to payment-service.
func (pc *paymentClientGRPC) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, constant.PaymentClientTimeout)
	defer cancel()

	req := connect.NewRequest(&pb.HealthRequest{})
	pkgconnect.AddAuthHeaders(ctx, req)

	resp, err := pc.client.Health(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Msg.GetStatus() != pb.HealthStatus_HEALTH_STATUS_SERVING {
		return fmt.Errorf("service unhealthy: %s", resp.Msg.GetStatus())
	}

	return nil
}
