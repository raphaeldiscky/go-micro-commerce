// Package routes provides the HTTP routes for the payment service.
package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/handler"
)

// SetupWebhookRoutes sets up webhook routes for payment gateways.
func SetupWebhookRoutes(e *echo.Echo, h *handler.WebhookHandler) {
	webhook := e.Group("/webhooks")

	// Stripe webhook endpoint
	webhook.POST("/stripe", h.HandleStripeWebhook)

	// Health check for webhook handler
	webhook.GET("/health", h.HealthCheck)
}
