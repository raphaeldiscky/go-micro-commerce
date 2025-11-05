// Package client provides external service clients for the payment service.
package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// GatewayClientStrategy defines the common interface for payment gateway service integration (Strategy Pattern).
type GatewayClientStrategy interface {
	// ProcessPayment processes a payment through the gateway
	ProcessPayment(
		ctx context.Context,
		req *dto.PaymentGatewayRequest,
	) (*dto.PaymentGatewayResponse, error)

	// GetPaymentStatus retrieves the status of a payment
	GetPaymentStatus(
		ctx context.Context,
		paymentID uuid.UUID,
		gatewayID string,
	) (*dto.PaymentGatewayResponse, error)

	// CapturePayment captures an authorized payment
	CapturePayment(
		ctx context.Context,
		paymentID uuid.UUID,
		gatewayID string,
		amount decimal.Decimal,
	) (*dto.PaymentGatewayResponse, error)

	// CancelPayment cancels or voids an authorized payment
	CancelPayment(ctx context.Context, gatewayID string) error

	// RefundPayment refunds a completed payment
	RefundPayment(ctx context.Context, req *dto.RefundRequest) (*dto.RefundResponse, error)

	// GetRefundStatus retrieves the status of a refund
	GetRefundStatus(
		ctx context.Context,
		refundID uuid.UUID,
		paymentID uuid.UUID,
		gatewayRefundID string,
	) (*dto.RefundResponse, error)
}
