// Package handler provides HTTP request handlers for the payment service.
package handler

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/stripe/stripe-go/v83"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
)

// WebhookHandler handles webhook events from payment gateways.
type WebhookHandler struct {
	config          *config.PaymentGatewayConfig
	logger          logger.Logger
	stripeClient    *stripe.Client
	processedEvents sync.Map // For idempotency: map[string]bool
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(
	cfg *config.PaymentGatewayConfig,
	appLogger logger.Logger,
	stripeClient *stripe.Client,
) *WebhookHandler {
	return &WebhookHandler{
		config:          cfg,
		logger:          appLogger,
		stripeClient:    stripeClient,
		processedEvents: sync.Map{},
	}
}

// HandleStripeWebhook handles Stripe webhook events using V2 thin events pattern.
func (h *WebhookHandler) HandleStripeWebhook(c echo.Context) error {
	const maxBodyBytes = int64(65536)

	// Read request body with size limit
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBodyBytes)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.Errorf("Error reading request body: %v", err)
		return err
	}

	// Parse and verify signature using V2 thin events
	signature := c.Request().Header.Get("Stripe-Signature")

	eventNotification, err := h.stripeClient.ParseEventNotification(
		payload,
		signature,
		h.config.StripeWebhookSecret,
	)
	if err != nil {
		h.logger.Errorf("Error verifying webhook signature: %v", err)
		return err
	}

	// Return 200 immediately before complex logic
	// Process event asynchronously in background
	go h.processEventNotification(context.Background(), eventNotification)

	return c.JSON(http.StatusOK, map[string]string{
		"received": "true",
	})
}

// processEventNotification processes Stripe event notifications asynchronously.
func (h *WebhookHandler) processEventNotification(
	ctx context.Context,
	eventNotification any,
) {
	// Get event type for logging
	eventType := ""
	if evt, ok := eventNotification.(*stripe.UnknownEventNotification); ok {
		eventType = evt.Type
	}

	h.logger.Infof("Processing Stripe event: %s", eventType)

	// Type switch on EventNotification to handle different event types
	switch evt := eventNotification.(type) {
	case *stripe.UnknownEventNotification:
		// PaymentIntent events are represented as UnknownEventNotification
		// since they might not be in the SDK yet
		h.handleUnknownEvent(ctx, evt)
	default:
		// Log unhandled event types
		h.logger.Infof("Unhandled event type: %s", eventType)
	}
}

// handleUnknownEvent handles UnknownEventNotification events (PaymentIntent events).
func (h *WebhookHandler) handleUnknownEvent(
	ctx context.Context,
	evt *stripe.UnknownEventNotification,
) {
	// Check idempotency - skip if already processed
	eventID := evt.ID
	if _, processed := h.processedEvents.LoadOrStore(eventID, true); processed {
		h.logger.Infof("Event %s already processed, skipping", eventID)
		return
	}

	// Match on event type for PaymentIntent events
	switch evt.Type {
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(ctx, evt)
	case "payment_intent.payment_failed":
		h.handlePaymentIntentFailed(ctx, evt)
	case "payment_intent.canceled":
		h.handlePaymentIntentCanceled(ctx, evt)
	case "charge.refunded":
		h.handleChargeRefunded(ctx, evt)
	default:
		h.logger.Infof("Skipping unknown event type: %s", evt.Type)
	}
}

// handlePaymentIntentSucceeded handles successful payment events.
func (h *WebhookHandler) handlePaymentIntentSucceeded(
	ctx context.Context,
	evt *stripe.UnknownEventNotification,
) {
	// Fetch the full event from Stripe API
	_, err := evt.FetchEvent(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching event: %v", err)
		return
	}

	h.logger.Infof("Payment succeeded event: %s", evt.ID)

	// Fetch the related PaymentIntent object
	relatedObj, err := evt.FetchRelatedObject(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching related object: %v", err)
		return
	}

	if relatedObj == nil {
		h.logger.Error("No related object found for payment_intent.succeeded event")
		return
	}

	h.logger.Infof(
		"Payment succeeded for PaymentIntent: %s",
		evt.Type,
	)

	// TODO: Update payment status in database
	// Note: The actual payment processing is done via ProcessPayment endpoint
	// This webhook serves as a confirmation and backup mechanism
}

// handlePaymentIntentFailed handles failed payment events.
func (h *WebhookHandler) handlePaymentIntentFailed(
	ctx context.Context,
	evt *stripe.UnknownEventNotification,
) {
	// Fetch the full event from Stripe API
	_, err := evt.FetchEvent(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching event: %v", err)
		return
	}

	h.logger.Infof("Payment failed event: %s", evt.ID)

	// Fetch the related PaymentIntent object
	relatedObj, err := evt.FetchRelatedObject(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching related object: %v", err)
		return
	}

	if relatedObj == nil {
		h.logger.Error("No related object found for payment_intent.payment_failed event")
		return
	}

	h.logger.Errorf(
		"Payment failed: %s",
		evt.Type,
	)

	// TODO: Update payment status to failed in database
	// TODO: Notify customer of payment failure
}

// handlePaymentIntentCanceled handles canceled payment events.
func (h *WebhookHandler) handlePaymentIntentCanceled(
	ctx context.Context,
	evt *stripe.UnknownEventNotification,
) {
	// Fetch the full event from Stripe API
	_, err := evt.FetchEvent(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching event: %v", err)
		return
	}

	h.logger.Infof("Payment canceled event: %s", evt.ID)

	// Fetch the related PaymentIntent object
	relatedObj, err := evt.FetchRelatedObject(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching related object: %v", err)
		return
	}

	if relatedObj == nil {
		h.logger.Error("No related object found for payment_intent.canceled event")
		return
	}

	h.logger.Infof(
		"Payment canceled: %s",
		evt.Type,
	)

	// TODO: Update payment status to canceled in database
}

// handleChargeRefunded handles refund events.
func (h *WebhookHandler) handleChargeRefunded(
	ctx context.Context,
	evt *stripe.UnknownEventNotification,
) {
	// Fetch the full event from Stripe API
	_, err := evt.FetchEvent(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching event: %v", err)
		return
	}

	h.logger.Infof("Charge refunded event: %s", evt.ID)

	// Fetch the related Charge object
	relatedObj, err := evt.FetchRelatedObject(ctx)
	if err != nil {
		h.logger.Errorf("Error fetching related object: %v", err)
		return
	}

	if relatedObj == nil {
		h.logger.Error("No related object found for charge.refunded event")
		return
	}

	h.logger.Infof(
		"Charge refunded: %s",
		evt.Type,
	)

	// TODO: Update refund status in database
}

// HealthCheck returns the health status of the webhook handler.
func (h *WebhookHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "webhook-handler",
	})
}
