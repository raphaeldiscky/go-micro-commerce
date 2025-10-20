// Package handler provides HTTP request handlers for the payment service.
package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// WebhookHandler handles webhook events from payment gateways.
type WebhookHandler struct {
	webhookService service.WebhookService
	logger         logger.Logger
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(
	webhookService service.WebhookService,
	appLogger logger.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         appLogger,
	}
}

// HandleStripeWebhook handles Stripe webhook events.
// This endpoint receives webhook events from Stripe after client-side payment confirmation.
// Events include: payment_intent.succeeded, payment_intent.failed, etc.
func (h *WebhookHandler) HandleStripeWebhook(c echo.Context) error {
	const maxBodyBytes = int64(65536)

	// Read request body with size limit to prevent DoS attacks
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBodyBytes)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.logger.Errorf("Error reading webhook request body: %v", err)

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Get Stripe signature from header for verification
	signature := c.Request().Header.Get(constant.StripeSignatureHeader)
	if signature == "" {
		h.logger.Error("Missing Stripe signature header")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing signature",
		})
	}

	// Create context with timeout for webhook processing
	ctx, cancel := context.WithTimeout(c.Request().Context(), constant.WebhookRequestTimeout)
	defer cancel()

	// Process webhook event (signature verification happens inside)
	if err = h.webhookService.HandleStripeWebhook(ctx, payload, signature); err != nil {
		h.logger.Errorf("Error processing webhook: %v", err)
		// Return 400 for signature verification errors, 500 for processing errors
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid webhook signature" {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, map[string]string{
			"error": err.Error(),
		})
	}

	h.logger.Info("Webhook processed successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}
