// Package service provides circuit breaker management for different services.
package service

import (
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"
	"github.com/sony/gobreaker"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
)

// CircuitBreakerService manages circuit breakers for different services.
type CircuitBreakerService struct {
	breakers  map[string]*gobreaker.CircuitBreaker
	mutex     sync.RWMutex
	logger    logger.Logger
	config    *config.Config
	telemetry *telemetry.Telemetry
}

// NewCircuitBreakerService creates a new circuit breaker service.
func NewCircuitBreakerService(
	appLogger logger.Logger,
	cfg *config.Config,
	tel *telemetry.Telemetry,
) *CircuitBreakerService {
	return &CircuitBreakerService{
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		logger:    appLogger,
		config:    cfg,
		telemetry: tel,
	}
}

// GetBreaker returns a circuit breaker for the given service.
func (cb *CircuitBreakerService) GetBreaker(serviceName string) *gobreaker.CircuitBreaker {
	cb.mutex.RLock()
	breaker, exists := cb.breakers[serviceName]
	cb.mutex.RUnlock()

	if exists {
		return breaker
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if breaker, exists = cb.breakers[serviceName]; exists {
		return breaker
	}

	// Create new circuit breaker
	settings := gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: cb.config.CircuitBreaker.MaxRequests,
		Interval:    cb.config.CircuitBreaker.Interval,
		Timeout:     cb.config.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)

			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cb.logger.Infof(
				"Circuit breaker state changed: service=%s from=%s to=%s",
				name,
				from.String(),
				to.String(),
			)

			// Record circuit breaker state change in metrics
			// State: 0=closed, 1=half-open, 2=open
			var stateValue float64
			switch to {
			case gobreaker.StateClosed:
				stateValue = 0
			case gobreaker.StateHalfOpen:
				stateValue = 1
			case gobreaker.StateOpen:
				stateValue = 2
			}
			if cb.telemetry != nil {
				cb.telemetry.SetCircuitBreakerState(name, stateValue)
			}
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	cb.breakers[serviceName] = breaker

	return breaker
}

// Execute executes a request through the circuit breaker.
func (cb *CircuitBreakerService) Execute(
	serviceName string,
	req func() (any, error),
) (any, error) {
	breaker := cb.GetBreaker(serviceName)

	return breaker.Execute(req)
}
