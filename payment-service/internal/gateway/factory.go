// Package gateway provides payment gateway implementations and factory.
package gateway

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/gateway/mock"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/gateway/stripe"
)

// Factory creates payment gateway clients based on configuration.
type Factory struct {
	config *config.PaymentGatewayConfig
	logger logger.Logger
}

// NewFactory creates a new gateway factory.
func NewFactory(cfg *config.PaymentGatewayConfig, appLogger logger.Logger) *Factory {
	return &Factory{
		config: cfg,
		logger: appLogger,
	}
}

// CreateGateways creates all available payment gateway clients.
func (f *Factory) CreateGateways() map[string]client.PaymentGatewayClient {
	f.logger.Info("Initializing all available payment gateway clients")

	gateways := make(map[string]client.PaymentGatewayClient)

	// Initialize Stripe gateway
	gateways["stripe"] = stripe.NewStripeClient(f.config, f.logger)
	f.logger.Info("Stripe payment gateway client initialized")

	// Initialize Mock gateway
	gateways["mock"] = mock.NewMockClient()

	f.logger.Info("Mock payment gateway client initialized")

	return gateways
}
