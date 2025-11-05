package provider

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

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
	// Initialize payment service
	paymentService := service.NewPaymentService(
		providers.DataStore,
		appLogger,
		providers.PaymentGatewayFactory,
	)
	providers.PaymentService = paymentService

	// Initialize payment handler and routes
	paymentHandler := handler.NewPaymentHandler(paymentService)
	routes.SetupPaymentRoutes(e, paymentHandler)

	// Initialize webhook service and handler
	webhookService := service.NewWebhookService(
		providers.DataStore,
		appLogger,
		cfg.PaymentGateway.StripeWebhookEndpointSecret,
	)
	webhookHandler := handler.NewWebhookHandler(webhookService, appLogger)
	routes.SetupWebhookRoutes(e, webhookHandler)

	graphResolver := SetupGraphQLResolver(paymentService)
	routes.SetupGraphQLRoutes(e, cfg, graphResolver, appLogger)
}
