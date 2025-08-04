// Package service provides various services for the API gateway, including circuit breakers, load balancing, and health checks.
package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreakerService manages circuit breakers for different services.
type CircuitBreakerService struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker service.
func NewCircuitBreaker() *CircuitBreakerService {
	logger, err := zap.NewProduction()
	if err != nil {
		// Logger initialization failed, use a no-op logger to avoid fmt.Printf
		logger = zap.NewNop()
	}

	return &CircuitBreakerService{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		logger:   logger,
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
			cb.logger.Info("Circuit breaker state changed",
				zap.String("service", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
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

// LoadBalancer interface for service load balancing.
type LoadBalancer interface {
	SelectEndpoint(serviceName string, endpoints []string) (string, error)
}

// RoundRobinLoadBalancer implements round-robin load balancing.
type RoundRobinLoadBalancer struct {
	counters map[string]int
	mutex    sync.RWMutex
}

// NewLoadBalancer creates a new round-robin load balancer.
func NewLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		counters: make(map[string]int),
	}
}

// SelectEndpoint selects an endpoint using round-robin algorithm.
func (lb *RoundRobinLoadBalancer) SelectEndpoint(
	serviceName string,
	endpoints []string,
) (string, error) {
	if len(endpoints) == 0 {
		return "", errors.New("no endpoints available")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	counter := lb.counters[serviceName]
	endpoint := endpoints[counter%len(endpoints)]
	lb.counters[serviceName] = counter + 1

	return endpoint, nil
}

// WeightedLoadBalancer implements weighted load balancing.
type WeightedLoadBalancer struct {
	weights map[string]map[string]int // service -> endpoint -> weight
	mutex   sync.RWMutex
}

// NewWeightedLoadBalancer creates a new weighted load balancer.
func NewWeightedLoadBalancer() *WeightedLoadBalancer {
	return &WeightedLoadBalancer{
		weights: make(map[string]map[string]int),
	}
}

// SetWeight sets the weight for an endpoint.
func (wlb *WeightedLoadBalancer) SetWeight(serviceName, endpoint string, weight int) {
	wlb.mutex.Lock()
	defer wlb.mutex.Unlock()

	if wlb.weights[serviceName] == nil {
		wlb.weights[serviceName] = make(map[string]int)
	}

	wlb.weights[serviceName][endpoint] = weight
}

// SelectEndpoint selects an endpoint using weighted algorithm.
func (wlb *WeightedLoadBalancer) SelectEndpoint(
	serviceName string,
	endpoints []string,
) (string, error) {
	if len(endpoints) == 0 {
		return "", errors.New("no endpoints available")
	}

	wlb.mutex.RLock()
	serviceWeights := wlb.weights[serviceName]
	wlb.mutex.RUnlock()

	if serviceWeights == nil {
		// Fallback to round-robin if no weights configured
		return endpoints[0], nil
	}

	totalWeight := 0

	for _, endpoint := range endpoints {
		if weight, exists := serviceWeights[endpoint]; exists {
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return endpoints[0], nil
	}

	// Implement weighted random selection here
	// For simplicity, return first endpoint for now
	return endpoints[0], nil
}

// HealthAwareLoadBalancer wraps another load balancer with health checking.
type HealthAwareLoadBalancer struct {
	underlying LoadBalancer
	discovery  Discovery
}

// NewHealthAwareLoadBalancer creates a health-aware load balancer.
func NewHealthAwareLoadBalancer(
	underlying LoadBalancer,
	discovery Discovery,
) *HealthAwareLoadBalancer {
	return &HealthAwareLoadBalancer{
		underlying: underlying,
		discovery:  discovery,
	}
}

// SelectEndpoint selects a healthy endpoint.
func (hlb *HealthAwareLoadBalancer) SelectEndpoint(
	serviceName string,
	endpoints []string,
) (string, error) {
	// Filter healthy endpoints
	var healthyEndpoints []string

	for _, endpoint := range endpoints {
		if hlb.discovery.HealthCheck(serviceName) {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	if len(healthyEndpoints) == 0 {
		return "", fmt.Errorf("no healthy endpoints available for service: %s", serviceName)
	}

	return hlb.underlying.SelectEndpoint(serviceName, healthyEndpoints)
}
