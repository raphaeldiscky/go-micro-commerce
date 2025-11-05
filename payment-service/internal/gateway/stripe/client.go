// Package stripe provides Stripe payment gateway integration.
package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"github.com/stripe/stripe-go/v83/refund"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// stripeClient implements PaymentGatewayClient for Stripe.
type stripeClient struct {
	logger logger.Logger
}

// NewStripeClient creates a new Stripe payment gateway client (Adapter Pattern).
func NewStripeClient(
	cfg *config.PaymentGatewayConfig,
	appLogger logger.Logger,
) client.GatewayClientStrategy {
	// Set global Stripe API key for SDK functions
	//nolint:reassign // Stripe SDK requires setting the global Key
	stripe.Key = cfg.StripeSecretKey

	return &stripeClient{
		logger: appLogger,
	}
}

const (
	multiplyAmount = 100
)

// ProcessPayment creates a Stripe PaymentIntent with the payment method attached.
// For PCI compliance, this does NOT confirm the payment - confirmation happens client-side
// with Stripe.js using the returned client_secret. This eliminates raw card data from our servers.
// Context is used for request timeout and cancellation.
func (c *stripeClient) ProcessPayment(
	ctx context.Context,
	req *dto.PaymentGatewayRequest,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Creating Stripe PaymentIntent for payment %s, amount: %s %s, payment_method: %s",
		req.PaymentID,
		req.Amount.String(),
		req.Currency,
		req.PaymentMethodID,
	)

	// Convert amount to smallest currency unit (cents for USD, yen for JPY, etc.)
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	metadata := map[string]string{
		"customer_id": req.CustomerID.String(),
	}

	// Add expiry timestamp if provided for 24-hour payment window tracking
	if req.ExpiresAt != nil {
		metadata["expires_at"] = req.ExpiresAt.Format(time.RFC3339)
	}

	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amountInCents),
		Currency:    stripe.String(req.Currency),
		Description: stripe.String(req.Description),
		Confirm: stripe.Bool(
			false,
		), // Client selects payment method and confirms with Stripe.js
		Metadata: metadata,
	}

	// Set customer email for receipts
	if req.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(req.CustomerEmail)
	}

	// Set idempotency key for safe retries
	params.IdempotencyKey = stripe.String(req.IdempotencyKey)

	// Enable automatic payment methods based on customer's location and currency
	// Stripe recommends NOT hardcoding payment_method_types - let Stripe choose
	// optimal methods based on user's location, wallets, and preferences
	params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
		Enabled:        stripe.Bool(true),
		AllowRedirects: stripe.String("always"), // Allow redirect-based methods (iDEAL, SEPA, etc.)
	}

	// Set context for timeout and cancellation
	params.Context = ctx

	// Create PaymentIntent (not confirmed yet - client will confirm)
	pi, err := paymentintent.New(params)
	if err != nil {
		c.logger.Errorf("Failed to create Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	c.logger.Infof(
		"Stripe PaymentIntent created: %s, status: %s, requires_action: %v",
		pi.ID,
		pi.Status,
		pi.Status == stripe.PaymentIntentStatusRequiresAction ||
			pi.Status == stripe.PaymentIntentStatusRequiresPaymentMethod,
	)

	return MapPaymentIntentToResponse(pi, req.PaymentID)
}

// GetPaymentStatus retrieves the status of a Stripe PaymentIntent.
// Context is used for request timeout and cancellation.
func (c *stripeClient) GetPaymentStatus(
	ctx context.Context,
	paymentID uuid.UUID,
	gatewayID string,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Retrieving Stripe payment status for: %s (payment ID: %s)",
		gatewayID,
		paymentID,
	)

	params := &stripe.PaymentIntentParams{}
	params.Context = ctx

	pi, err := paymentintent.Get(gatewayID, params)
	if err != nil {
		c.logger.Errorf("Failed to retrieve Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	return MapPaymentIntentToResponse(pi, paymentID)
}

// CapturePayment captures an authorized Stripe PaymentIntent.
// Context is used for request timeout and cancellation.
func (c *stripeClient) CapturePayment(
	ctx context.Context,
	paymentID uuid.UUID,
	gatewayID string,
	amount decimal.Decimal,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Capturing Stripe payment: %s, amount: %s (payment ID: %s)",
		gatewayID,
		amount.String(),
		paymentID,
	)

	// Convert amount to smallest currency unit
	amountInCents := amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.PaymentIntentCaptureParams{
		AmountToCapture: stripe.Int64(amountInCents),
	}
	params.Context = ctx

	pi, err := paymentintent.Capture(gatewayID, params)
	if err != nil {
		c.logger.Errorf("Failed to capture Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to capture payment intent: %w", err)
	}

	return MapPaymentIntentToResponse(pi, paymentID)
}

// CancelPayment cancels a Stripe PaymentIntent.
// Context is used for request timeout and cancellation.
func (c *stripeClient) CancelPayment(ctx context.Context, gatewayID string) error {
	c.logger.Infof("Canceling Stripe payment: %s", gatewayID)

	params := &stripe.PaymentIntentCancelParams{}
	params.Context = ctx

	_, err := paymentintent.Cancel(gatewayID, params)
	if err != nil {
		c.logger.Errorf("Failed to cancel Stripe PaymentIntent: %v", err)
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	c.logger.Infof("Stripe payment canceled successfully: %s", gatewayID)

	return nil
}

// RefundPayment creates a refund for a Stripe charge.
// Context is used for request timeout and cancellation.
func (c *stripeClient) RefundPayment(
	ctx context.Context,
	req *dto.RefundRequest,
) (*dto.RefundResponse, error) {
	c.logger.Infof(
		"Creating Stripe refund for: %s, amount: %s %s",
		req.PaymentIntentID,
		req.Amount.String(),
		req.Currency,
	)

	// Convert amount to smallest currency unit
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentIntentID),
		Amount:        stripe.Int64(amountInCents),
		Reason:        stripe.String(stripe.RefundReasonRequestedByCustomer),
		Metadata: map[string]string{
			"refund_id": req.RefundID.String(),
		},
	}
	params.Context = ctx

	if req.Reason != "" {
		params.Metadata["reason"] = req.Reason
	}

	r, err := refund.New(params)
	if err != nil {
		c.logger.Errorf("Failed to create Stripe refund: %v", err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	c.logger.Infof("Stripe refund created: %s, status: %s", r.ID, r.Status)

	return MapRefundToResponse(r, req.RefundID, req.PaymentID)
}

// GetRefundStatus retrieves the status of a Stripe refund.
// Context is used for request timeout and cancellation.
func (c *stripeClient) GetRefundStatus(
	ctx context.Context,
	refundID uuid.UUID,
	paymentID uuid.UUID,
	gatewayRefundID string,
) (*dto.RefundResponse, error) {
	c.logger.Infof(
		"Retrieving Stripe refund status for: %s (refund ID: %s, payment ID: %s)",
		gatewayRefundID,
		refundID,
		paymentID,
	)

	params := &stripe.RefundParams{}
	params.Context = ctx

	r, err := refund.Get(gatewayRefundID, params)
	if err != nil {
		c.logger.Errorf("Failed to retrieve Stripe refund: %v", err)
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return MapRefundToResponse(r, refundID, paymentID)
}
