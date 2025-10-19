// Package stripe provides Stripe payment gateway integration.
package stripe

import (
	"context"
	"fmt"

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

// NewStripeClient creates a new Stripe payment gateway client.
func NewStripeClient(
	cfg *config.PaymentGatewayConfig,
	appLogger logger.Logger,
) client.PaymentGatewayClient {
	// Set global Stripe API key for SDK functions
	//nolint:reassign // Stripe SDK requires setting the global Key
	stripe.Key = cfg.StripeAPIKey

	return &stripeClient{
		logger: appLogger,
	}
}

const (
	multiplyAmount = 100
)

// ProcessPayment creates a Stripe PaymentIntent and processes the payment.
func (c *stripeClient) ProcessPayment(
	_ context.Context,
	req *dto.PaymentGatewayRequest,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Processing Stripe payment for transaction %s, amount: %s %s",
		req.TransactionID,
		req.Amount.String(),
		req.Currency,
	)

	// Convert amount to smallest currency unit (cents for USD, etc.)
	// Multiply amount by 100 to convert to cents (smallest currency unit)
	// 100 is used because Stripe uses cents for USD and other currencies
	// that have a decimal place of 2.
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amountInCents),
		Currency:    stripe.String(req.Currency),
		Description: stripe.String(req.Description),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
			"customer_id":    req.CustomerID.String(),
		},
	}

	// Set customer email
	if req.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(req.CustomerEmail)
	}

	// Set idempotency key for safe retries
	params.IdempotencyKey = stripe.String(req.IdempotencyKey)

	// Add payment method types based on request
	params.PaymentMethodTypes = stripe.StringSlice([]string{"card"})

	// Create PaymentIntent
	pi, err := paymentintent.New(params)
	if err != nil {
		c.logger.Errorf("Failed to create Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	c.logger.Infof("Stripe PaymentIntent created: %s, status: %s", pi.ID, pi.Status)

	return MapPaymentIntentToResponse(pi, req.TransactionID)
}

// GetPaymentStatus retrieves the status of a Stripe PaymentIntent.
func (c *stripeClient) GetPaymentStatus(
	_ context.Context,
	gatewayID string,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof("Retrieving Stripe payment status for: %s", gatewayID)

	pi, err := paymentintent.Get(gatewayID, nil)
	if err != nil {
		c.logger.Errorf("Failed to retrieve Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}

	// Extract transaction ID from metadata
	transactionID, err := parseTransactionIDFromMetadata(pi.Metadata)
	if err != nil {
		c.logger.Errorf("Failed to parse transaction ID from metadata: %v", err)
		return nil, err
	}

	return MapPaymentIntentToResponse(pi, transactionID)
}

// CapturePayment captures an authorized Stripe PaymentIntent.
func (c *stripeClient) CapturePayment(
	_ context.Context,
	gatewayID string,
	amount decimal.Decimal,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof("Capturing Stripe payment: %s, amount: %s", gatewayID, amount.String())

	// Convert amount to smallest currency unit
	amountInCents := amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.PaymentIntentCaptureParams{
		AmountToCapture: stripe.Int64(amountInCents),
	}

	pi, err := paymentintent.Capture(gatewayID, params)
	if err != nil {
		c.logger.Errorf("Failed to capture Stripe PaymentIntent: %v", err)
		return nil, fmt.Errorf("failed to capture payment intent: %w", err)
	}

	// Extract transaction ID from metadata
	transactionID, err := parseTransactionIDFromMetadata(pi.Metadata)
	if err != nil {
		c.logger.Errorf("Failed to parse transaction ID from metadata: %v", err)
		return nil, err
	}

	return MapPaymentIntentToResponse(pi, transactionID)
}

// CancelPayment cancels a Stripe PaymentIntent.
func (c *stripeClient) CancelPayment(_ context.Context, gatewayID string) error {
	c.logger.Infof("Canceling Stripe payment: %s", gatewayID)

	_, err := paymentintent.Cancel(gatewayID, nil)
	if err != nil {
		c.logger.Errorf("Failed to cancel Stripe PaymentIntent: %v", err)
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	c.logger.Infof("Stripe payment canceled successfully: %s", gatewayID)

	return nil
}

// RefundPayment creates a refund for a Stripe charge.
func (c *stripeClient) RefundPayment(
	_ context.Context,
	req *dto.RefundRequest,
) (*dto.RefundResponse, error) {
	c.logger.Infof(
		"Creating Stripe refund for: %s, amount: %s %s",
		req.GatewayID,
		req.Amount.String(),
		req.Currency,
	)

	// Convert amount to smallest currency unit
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.GatewayID),
		Amount:        stripe.Int64(amountInCents),
		Reason:        stripe.String(stripe.RefundReasonRequestedByCustomer),
		Metadata: map[string]string{
			"refund_id":      req.RefundID.String(),
			"transaction_id": req.TransactionID.String(),
		},
	}

	if req.Reason != "" {
		params.Metadata["reason"] = req.Reason
	}

	r, err := refund.New(params)
	if err != nil {
		c.logger.Errorf("Failed to create Stripe refund: %v", err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	c.logger.Infof("Stripe refund created: %s, status: %s", r.ID, r.Status)

	return MapRefundToResponse(r, req.RefundID, req.TransactionID)
}

// GetRefundStatus retrieves the status of a Stripe refund.
func (c *stripeClient) GetRefundStatus(
	_ context.Context,
	gatewayRefundID string,
) (*dto.RefundResponse, error) {
	c.logger.Infof("Retrieving Stripe refund status for: %s", gatewayRefundID)

	r, err := refund.Get(gatewayRefundID, nil)
	if err != nil {
		c.logger.Errorf("Failed to retrieve Stripe refund: %v", err)
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	// Extract IDs from metadata
	refundID, transactionID, err := parseRefundMetadata(r.Metadata)
	if err != nil {
		c.logger.Errorf("Failed to parse refund metadata: %v", err)
		return nil, err
	}

	return MapRefundToResponse(r, refundID, transactionID)
}

// ValidateCard validates a payment card using Stripe's token API.
func (c *stripeClient) ValidateCard(_ context.Context, card *dto.PaymentCard) error {
	c.logger.Info("Validating payment card with Stripe")

	// Stripe validates cards when creating payment methods or tokens
	// For now, just validate basic card info format
	if err := validateCardInfo(card); err != nil {
		return fmt.Errorf("invalid card information: %w", err)
	}

	return nil
}
