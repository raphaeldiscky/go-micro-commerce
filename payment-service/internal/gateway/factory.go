// Package gateway provides payment gateway implementations and factory.
package gateway

import (
	"fmt"

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

// CreateGateway creates a payment gateway client based on the provider type.
func (f *Factory) CreateGateway(provider string) (client.PaymentGatewayClient, error) {
	switch provider {
	case "stripe":
		f.logger.Info("Creating Stripe payment gateway client")
		return stripe.NewStripeClient(f.config, f.logger), nil
	case "mock":
		f.logger.Info("Creating mock payment gateway client")
		return mock.NewMockClient(), nil
	default:
		return nil, fmt.Errorf("unsupported payment gateway provider: %s", provider)
	}
}
