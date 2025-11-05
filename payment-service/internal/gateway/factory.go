// Package gateway provides payment gateway implementations and factory.
package gateway

import (
	"fmt"
	"sync"

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
	mu     sync.RWMutex
	cache  map[string]client.GatewayClientStrategy
}

// NewFactory creates a new gateway factory (Factory Pattern).
func NewFactory(cfg *config.PaymentGatewayConfig, appLogger logger.Logger) *Factory {
	return &Factory{
		config: cfg,
		logger: appLogger,
		cache:  make(map[string]client.GatewayClientStrategy),
	}
}

// GetGateway returns a payment gateway client for the specified provider.
// Gateways are lazily initialized and cached for subsequent requests.
func (f *Factory) GetGateway(provider string) (client.GatewayClientStrategy, error) {
	// Check cache first with read lock
	f.mu.RLock()

	if gateway, exists := f.cache[provider]; exists {
		f.mu.RUnlock()
		return gateway, nil
	}

	f.mu.RUnlock()

	// Create gateway with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check in case another goroutine created it
	if gateway, exists := f.cache[provider]; exists {
		return gateway, nil
	}

	// Create gateway based on provider
	var gateway client.GatewayClientStrategy

	switch provider {
	case "stripe":
		gateway = stripe.NewStripeClient(f.config, f.logger)
		f.logger.Info("Stripe payment gateway client initialized")
	case "mock":
		gateway = mock.NewMockClient()

		f.logger.Info("Mock payment gateway client initialized")
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}

	// Cache the gateway
	f.cache[provider] = gateway

	return gateway, nil
}
