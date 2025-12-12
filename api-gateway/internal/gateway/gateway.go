// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

// Gateway represents the API Gateway.
type Gateway struct {
	logger           logger.Logger
	serviceDiscovery service.Discovery
	circuitBreaker   *service.CircuitBreakerService
	telemetry        *telemetry.Telemetry
	config           *config.Config
}

// Config holds gateway configuration.
type Config struct {
	Logger           logger.Logger
	ServiceDiscovery service.Discovery
	CircuitBreaker   *service.CircuitBreakerService
	Telemetry        *telemetry.Telemetry
	Config           *config.Config
}

// NewAPIGateway creates a new API Gateway instance.
func NewAPIGateway(cfg Config) *Gateway {
	return &Gateway{
		logger:           cfg.Logger,
		serviceDiscovery: cfg.ServiceDiscovery,
		circuitBreaker:   cfg.CircuitBreaker,
		telemetry:        cfg.Telemetry,
		config:           cfg.Config,
	}
}
