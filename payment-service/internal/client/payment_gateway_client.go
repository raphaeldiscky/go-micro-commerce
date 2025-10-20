// Package client provides external service clients for the payment service.
package client

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// PaymentGatewayClient defines the interface for payment gateway service integration.
type PaymentGatewayClient interface {
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

	// CreateSetupIntent creates a SetupIntent for collecting payment method without charging
	// Used for delayed payment confirmation pattern (save now, charge later)
	CreateSetupIntent(
		ctx context.Context,
		req *dto.SetupIntentRequest,
	) (*dto.SetupIntentResponse, error)

	// ChargeOffSession charges a saved payment method without customer present
	// Used after SetupIntent to charge the card when order status changes
	ChargeOffSession(
		ctx context.Context,
		req *dto.ChargeOffSessionRequest,
	) (*dto.PaymentGatewayResponse, error)

	// CreateOrRetrieveCustomer ensures a Stripe Customer exists for the given customer ID
	// Returns the Stripe Customer ID (cus_xxx)
	CreateOrRetrieveCustomer(
		ctx context.Context,
		customerID string,
		email string,
	) (string, error)
}
