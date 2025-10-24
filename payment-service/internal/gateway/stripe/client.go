// Package stripe provides Stripe payment gateway integration.
package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"github.com/stripe/stripe-go/v83/refund"
	"github.com/stripe/stripe-go/v83/setupintent"

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

// ProcessPayment creates a Stripe PaymentIntent with the payment method attached.
// For PCI compliance, this does NOT confirm the payment - confirmation happens client-side
// with Stripe.js using the returned client_secret. This eliminates raw card data from our servers.
func (c *stripeClient) ProcessPayment(
	_ context.Context,
	req *dto.PaymentGatewayRequest,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Creating Stripe PaymentIntent for transaction %s, amount: %s %s, payment_method: %s",
		req.TransactionID,
		req.Amount.String(),
		req.Currency,
		req.PaymentMethodID,
	)

	// Convert amount to smallest currency unit (cents for USD, yen for JPY, etc.)
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	metadata := map[string]string{
		"transaction_id": req.TransactionID.String(),
		"customer_id":    req.CustomerID.String(),
	}

	// Add expiry timestamp if provided for 24-hour payment window tracking
	if req.ExpiresAt != nil {
		metadata["expires_at"] = req.ExpiresAt.Format(time.RFC3339)
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountInCents),
		Currency:      stripe.String(req.Currency),
		Description:   stripe.String(req.Description),
		PaymentMethod: stripe.String(req.PaymentMethodID), // PM ID tokenized client-side
		Confirm:       stripe.Bool(false),                 // Client confirms with Stripe.js
		Metadata:      metadata,
	}

	// Set customer email for receipts
	if req.CustomerEmail != "" {
		params.ReceiptEmail = stripe.String(req.CustomerEmail)
	}

	// Set idempotency key for safe retries
	params.IdempotencyKey = stripe.String(req.IdempotencyKey)

	// Support multiple payment method types (cards, wallets, regional methods)
	params.PaymentMethodTypes = stripe.StringSlice([]string{
		"card",
		"link",       // Stripe Link
		"apple_pay",  // Apple Pay
		"google_pay", // Google Pay
	})

	// Enable automatic payment methods based on customer's location and currency
	params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
		Enabled:        stripe.Bool(true),
		AllowRedirects: stripe.String("always"), // Allow redirect-based methods (iDEAL, etc.)
	}

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

// CreateSetupIntent creates a SetupIntent for collecting payment method without charging.
// Used for delayed payment confirmation pattern (save now, charge later).
func (c *stripeClient) CreateSetupIntent(
	ctx context.Context,
	req *dto.SetupIntentRequest,
) (*dto.SetupIntentResponse, error) {
	c.logger.Infof("Creating SetupIntent for customer: %s, order: %s", req.CustomerID, req.OrderID)

	// 1. Create or retrieve Stripe Customer
	customerID, err := c.CreateOrRetrieveCustomer(ctx, req.CustomerID.String(), req.CustomerEmail)
	if err != nil {
		c.logger.Errorf("Failed to create/retrieve customer: %v", err)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// 2. Create SetupIntent
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
			"link",       // Stripe Link
			"apple_pay",  // Apple Pay (if available)
			"google_pay", // Google Pay (if available)
		}),
		Usage: stripe.String("off_session"), // Critical: allows charging without customer present
		Metadata: map[string]string{
			"order_id":    req.OrderID.String(),
			"customer_id": req.CustomerID.String(),
		},
	}

	si, err := setupintent.New(params)
	if err != nil {
		c.logger.Errorf("Failed to create SetupIntent: %v", err)
		return nil, fmt.Errorf("failed to create setup intent: %w", err)
	}

	c.logger.Infof("SetupIntent created: %s for customer: %s", si.ID, customerID)

	return &dto.SetupIntentResponse{
		SetupIntentID:    si.ID,
		ClientSecret:     si.ClientSecret,
		StripeCustomerID: customerID,
	}, nil
}

// ChargeOffSession charges a saved payment method without customer present.
// Used for delayed payment confirmation when customer already provided payment details.
func (c *stripeClient) ChargeOffSession(
	_ context.Context,
	req *dto.ChargeOffSessionRequest,
) (*dto.PaymentGatewayResponse, error) {
	c.logger.Infof(
		"Charging off-session: PM=%s, Customer=%s, Amount=%s %s",
		req.PaymentMethodID,
		req.StripeCustomerID,
		req.Amount.String(),
		req.Currency,
	)

	// Convert amount to smallest currency unit (cents)
	amountInCents := req.Amount.Mul(decimal.NewFromInt(multiplyAmount)).IntPart()

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountInCents),
		Currency:      stripe.String(req.Currency),
		Customer:      stripe.String(req.StripeCustomerID),
		PaymentMethod: stripe.String(req.PaymentMethodID),
		Confirm:       stripe.Bool(true), // Charge immediately on creation
		OffSession:    stripe.Bool(true), // Critical: charge without customer present
		Description:   stripe.String(req.Description),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
			"order_id":       req.OrderID.String(),
		},
	}

	// Create and confirm PaymentIntent in one call
	pi, err := paymentintent.New(params)
	if err != nil {
		c.logger.Errorf("Off-session charge failed: %v", err)
		return nil, fmt.Errorf("failed to charge off-session: %w", err)
	}

	c.logger.Infof(
		"Off-session charge successful: %s, status: %s",
		pi.ID,
		pi.Status,
	)

	return MapPaymentIntentToResponse(pi, req.TransactionID)
}

// CreateOrRetrieveCustomer ensures a Stripe Customer exists for the given customer ID.
// Searches for existing customer by metadata, creates new one if not found.
func (c *stripeClient) CreateOrRetrieveCustomer(
	_ context.Context,
	customerID string,
	email string,
) (string, error) {
	c.logger.Infof("Creating/retrieving Stripe Customer for: %s (%s)", customerID, email)

	// Search for existing customer by metadata
	searchParams := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("metadata['customer_id']:'%s'", customerID),
		},
	}

	iter := customer.Search(searchParams)
	if iter.Next() {
		cust := iter.Customer()
		c.logger.Infof("Found existing Stripe Customer: %s", cust.ID)

		return cust.ID, nil
	}

	// Check for search errors
	if iter.Err() != nil {
		c.logger.Warnf("Error searching for customer: %v", iter.Err())
		// Continue to create new customer
	}

	// Create new customer
	c.logger.Infof("Creating new Stripe Customer for: %s", customerID)

	createParams := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"customer_id": customerID,
		},
	}

	cust, err := customer.New(createParams)
	if err != nil {
		c.logger.Errorf("Failed to create Stripe Customer: %v", err)
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	c.logger.Infof("Created new Stripe Customer: %s for customer_id: %s", cust.ID, customerID)

	return cust.ID, nil
}
