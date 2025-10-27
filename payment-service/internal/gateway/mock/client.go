// Package mock provides a mock payment gateway client for testing.
package mock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

const (
	fakeGatewayDelay = time.Millisecond * 100
	fakeGatewayFee   = 0.029 // 2.9% fee
	fakeAmount       = 500000
)

// mockClient provides a simple mock implementation of PaymentGatewayClient.
type mockClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewMockClient creates a new instance of mockClient with test utilities.
func NewMockClient() client.PaymentGatewayClient {
	return &mockClient{
		shouldFail: false,
		delay:      fakeGatewayDelay,
	}
}

// ProcessPayment simulates creating a PaymentIntent with payment method attached.
// Mimics Stripe's behavior: returns PaymentIntent in requires_payment_method status
// with a client_secret for frontend confirmation.
func (c *mockClient) ProcessPayment(
	_ context.Context,
	req *dto.PaymentGatewayRequest,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, errors.New("payment gateway error")
	}

	// Generate fake gateway ID (pi_xxx format like Stripe)
	gatewayID := "pi_mock_" + uuid.NewString()

	// Generate fake client_secret (pi_xxx_secret_xxx format like Stripe)
	clientSecret := gatewayID + "_secret_" + uuid.NewString()[:16]

	// Check if payment method ID is present
	if req.PaymentMethodID == "" {
		return &dto.PaymentGatewayResponse{
			PaymentID:       req.PaymentID,
			GatewayID:       gatewayID,
			Status:          constant.PaymentGatewayStatusPending,
			Amount:          req.Amount,
			Currency:        req.Currency,
			ProcessedAt:     time.Now(),
			ClientSecret:    &clientSecret,
			RequiresAction:  true,
			GatewayResponse: map[string]any{"error": "payment_method required"},
		}, nil
	}

	// Simulate payment intent status (not confirmed yet - waiting for client)
	status := constant.PaymentGatewayStatusPending
	requiresAction := req.Amount.GreaterThan(decimal.NewFromInt(fakeAmount))

	// Simulate 3DS for amounts > 500k IDR

	fees := req.Amount.Mul(decimal.NewFromFloat(fakeGatewayFee)) // 2.9% fee

	return &dto.PaymentGatewayResponse{
		PaymentID:      req.PaymentID,
		GatewayID:      gatewayID,
		Status:         status,
		Amount:         req.Amount,
		Currency:       req.Currency,
		ProcessedAt:    time.Now(),
		ClientSecret:   &clientSecret, // For client-side confirmation
		RequiresAction: requiresAction,
		Fees:           &fees,
		GatewayResponse: map[string]any{
			"status":          string(status),
			"payment_method":  req.PaymentMethodID,
			"requires_action": requiresAction,
		},
	}, nil
}

// GetPaymentStatus retrieves payment status.
func (c *mockClient) GetPaymentStatus(
	_ context.Context,
	paymentID uuid.UUID,
	gatewayID string,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	return &dto.PaymentGatewayResponse{
		PaymentID:   paymentID,
		GatewayID:   gatewayID,
		Status:      constant.PaymentGatewayStatusSucceeded,
		ProcessedAt: time.Now(),
	}, nil
}

// CapturePayment captures an authorized payment.
func (c *mockClient) CapturePayment(
	_ context.Context,
	paymentID uuid.UUID,
	gatewayID string,
	amount decimal.Decimal,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	return &dto.PaymentGatewayResponse{
		PaymentID:   paymentID,
		GatewayID:   gatewayID,
		Status:      constant.PaymentGatewayStatusSucceeded,
		Amount:      amount,
		ProcessedAt: time.Now(),
	}, nil
}

// CancelPayment cancels a payment.
func (c *mockClient) CancelPayment(_ context.Context, _ string) error {
	time.Sleep(c.delay)

	return nil
}

// RefundPayment refunds a payment.
func (c *mockClient) RefundPayment(
	_ context.Context,
	req *dto.RefundRequest,
) (*dto.RefundResponse, error) {
	time.Sleep(c.delay)

	return &dto.RefundResponse{
		RefundID:        req.RefundID,
		PaymentID:       req.PaymentID,
		GatewayRefundID: uuid.NewString(),
		Status:          constant.RefundStatusSucceeded,
		Amount:          req.Amount,
		Currency:        req.Currency,
		ProcessedAt:     time.Now(),
	}, nil
}

// GetRefundStatus retrieves refund status.
func (c *mockClient) GetRefundStatus(
	_ context.Context,
	refundID uuid.UUID,
	paymentID uuid.UUID,
	gatewayRefundID string,
) (*dto.RefundResponse, error) {
	time.Sleep(c.delay)

	return &dto.RefundResponse{
		RefundID:        refundID,
		PaymentID:       paymentID,
		GatewayRefundID: gatewayRefundID,
		Status:          constant.RefundStatusSucceeded,
		ProcessedAt:     time.Now(),
	}, nil
}

// CreateOrRetrieveCustomer simulates creating or retrieving a Stripe customer.
func (c *mockClient) CreateOrRetrieveCustomer(
	_ context.Context,
	_ string,
	_ string,
) (string, error) {
	time.Sleep(c.delay)

	// Always return a new mock customer ID for simplicity
	return "cus_mock_" + uuid.NewString(), nil
}
