package service

import (
	"sync"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/sony/gobreaker"
)

// CircuitBreakerService manages circuit breakers for different services.
type CircuitBreakerService struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mutex    sync.RWMutex
	logger   logger.Logger
}

// NewCircuitBreakerService creates a new circuit breaker service.
func NewCircuitBreakerService(appLogger logger.Logger) *CircuitBreakerService {
	return &CircuitBreakerService{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		logger:   appLogger,
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

	// Double-check pattern
	if breaker, exists := cb.breakers[serviceName]; exists {
		return breaker
	}

	// Create new circuit breaker
	settings := gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
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
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	cb.breakers[serviceName] = breaker

	return breaker
}

// Execute executes a request through the circuit breaker.
func (cb *CircuitBreakerService) Execute(
	serviceName string,
	req func() (interface{}, error),
) (interface{}, error) {
	breaker := cb.GetBreaker(serviceName)

	return breaker.Execute(req)
}
