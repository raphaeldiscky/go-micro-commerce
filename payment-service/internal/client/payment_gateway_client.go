package client

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// PaymentGatewayClientInterface defines the interface for payment gateway service integration.
type PaymentGatewayClientInterface interface {
	// ProcessPayment processes a payment through the gateway
	ProcessPayment(
		ctx context.Context,
		req *dto.PaymentGatewayRequest,
	) (*dto.PaymentGatewayResponse, error)

	// GetPaymentStatus retrieves the status of a payment
	GetPaymentStatus(ctx context.Context, gatewayID string) (*dto.PaymentGatewayResponse, error)

	// CapturePayment captures an authorized payment
	CapturePayment(
		ctx context.Context,
		gatewayID string,
		amount decimal.Decimal,
	) (*dto.PaymentGatewayResponse, error)

	// CancelPayment cancels or voids an authorized payment
	CancelPayment(ctx context.Context, gatewayID string) error

	// RefundPayment refunds a completed payment
	RefundPayment(ctx context.Context, req *dto.RefundRequest) (*dto.RefundResponse, error)

	// GetRefundStatus retrieves the status of a refund
	GetRefundStatus(ctx context.Context, gatewayRefundID string) (*dto.RefundResponse, error)

	// ValidateCard validates a payment card without charging
	ValidateCard(ctx context.Context, card *dto.PaymentCard) error
}
