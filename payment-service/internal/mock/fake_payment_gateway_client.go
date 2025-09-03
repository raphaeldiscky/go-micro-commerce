package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/random"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// FakePaymentGatewayClient provides a simple mock implementation of PaymentGatewayClientInterface.
type FakePaymentGatewayClient struct {
	shouldFail bool
	delay      time.Duration
}

// NewFakePaymentGatewayClient creates a new instance of FakePaymentGatewayClient.
func NewFakePaymentGatewayClient() *FakePaymentGatewayClient {
	return &FakePaymentGatewayClient{
		shouldFail: false,
		delay:      time.Millisecond * 100,
	}
}

// SetShouldFail configures the client to simulate failures.
func (c *FakePaymentGatewayClient) SetShouldFail(shouldFail bool) {
	c.shouldFail = shouldFail
}

// ProcessPayment processes a payment through the gateway.
func (c *FakePaymentGatewayClient) ProcessPayment(
	_ context.Context,
	req *dto.PaymentGatewayRequest,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	if c.shouldFail {
		return nil, fmt.Errorf("payment gateway error")
	}

	// Simple success/failure logic
	status := constant.PaymentGatewayStatusSucceeded
	if req.Amount.GreaterThan(decimal.NewFromInt(1000000)) { // Fail payments > 1M IDR
		status = constant.PaymentGatewayStatusFailed
	}

	gatewayID := fmt.Sprintf("pi_%s", random.String(8))
	fees := req.Amount.Mul(decimal.NewFromFloat(0.029)) // 2.9% fee

	return &dto.PaymentGatewayResponse{
		TransactionID:   req.TransactionID,
		GatewayID:       gatewayID,
		Status:          status,
		Amount:          req.Amount,
		Currency:        req.Currency,
		ProcessedAt:     time.Now(),
		Fees:            &fees,
		GatewayResponse: map[string]interface{}{"status": string(status)},
	}, nil
}

// GetPaymentStatus retrieves payment status.
func (c *FakePaymentGatewayClient) GetPaymentStatus(
	_ context.Context,
	gatewayID string,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	return &dto.PaymentGatewayResponse{
		GatewayID:   gatewayID,
		Status:      constant.PaymentGatewayStatusSucceeded,
		ProcessedAt: time.Now(),
	}, nil
}

// CapturePayment captures an authorized payment.
func (c *FakePaymentGatewayClient) CapturePayment(
	_ context.Context,
	gatewayID string,
	amount decimal.Decimal,
) (*dto.PaymentGatewayResponse, error) {
	time.Sleep(c.delay)

	return &dto.PaymentGatewayResponse{
		GatewayID:   gatewayID,
		Status:      constant.PaymentGatewayStatusSucceeded,
		Amount:      amount,
		ProcessedAt: time.Now(),
	}, nil
}

// CancelPayment cancels a payment.
func (c *FakePaymentGatewayClient) CancelPayment(_ context.Context, _ string) error {
	time.Sleep(c.delay)

	return nil
}

// RefundPayment refunds a payment.
func (c *FakePaymentGatewayClient) RefundPayment(
	_ context.Context,
	req *dto.RefundRequest,
) (*dto.RefundResponse, error) {
	time.Sleep(c.delay)

	return &dto.RefundResponse{
		RefundID:        req.RefundID,
		TransactionID:   req.TransactionID,
		GatewayRefundID: fmt.Sprintf("re_%s", random.String(8)),
		Status:          constant.RefundStatusSucceeded,
		Amount:          req.Amount,
		Currency:        req.Currency,
		ProcessedAt:     time.Now(),
	}, nil
}

// GetRefundStatus retrieves refund status.
func (c *FakePaymentGatewayClient) GetRefundStatus(
	_ context.Context,
	gatewayRefundID string,
) (*dto.RefundResponse, error) {
	time.Sleep(c.delay)

	return &dto.RefundResponse{
		GatewayRefundID: gatewayRefundID,
		Status:          constant.RefundStatusSucceeded,
		ProcessedAt:     time.Now(),
	}, nil
}

// ValidateCard validates a payment card.
func (c *FakePaymentGatewayClient) ValidateCard(_ context.Context, _ *dto.PaymentCard) error {
	time.Sleep(c.delay)

	return nil
}
