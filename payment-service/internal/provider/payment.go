package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/stripe/stripe-go/v83"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/routes"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// SetupPayment initializes the payment-related routes and services.
func SetupPayment(
	e *echo.Echo,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) {
	paymentService := service.NewPaymentService(
		providers.DataStore,
		appLogger,
		providers.PaymentGatewayClient,
	)
	providers.PaymentService = paymentService
	paymentHandler := handler.NewPaymentHandler(paymentService)

	routes.SetupPaymentRoutes(e, paymentHandler)

	// Setup webhook handler and routes
	// Create Stripe client for webhook signature verification and event fetching
	stripeClient := stripe.NewClient(cfg.PaymentGateway.StripeAPIKey)
	webhookHandler := handler.NewWebhookHandler(
		cfg.PaymentGateway,
		appLogger,
		stripeClient,
	)
	routes.SetupWebhookRoutes(e, webhookHandler)
}
