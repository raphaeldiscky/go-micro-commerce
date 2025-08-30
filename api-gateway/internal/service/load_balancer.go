// Package service provides various services for the API gateway, including circuit breakers, load balancing, and health checks.
package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// LoadBalancer interface for service load balancing.
type LoadBalancer interface {
	SelectEndpoint(serviceName string, endpoints []string) (string, error)
}

// RoundRobinLoadBalancer implements round-robin load balancing.
type RoundRobinLoadBalancer struct {
	counters map[string]int
	mutex    sync.RWMutex
	logger   logger.Logger
}

// NewLoadBalancerService creates a new round-robin load balancer.
func NewLoadBalancerService(appLogger logger.Logger) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		counters: make(map[string]int),
		logger:   appLogger,
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

// HealthAwareLoadBalancerService wraps another load balancer with health checking.
type HealthAwareLoadBalancerService struct {
	underlying LoadBalancer
	discovery  Discovery
}

// NewHealthAwareLoadBalancerService creates a health-aware load balancer.
func NewHealthAwareLoadBalancerService(
	underlying LoadBalancer,
	discovery Discovery,
) *HealthAwareLoadBalancerService {
	return &HealthAwareLoadBalancerService{
		underlying: underlying,
		discovery:  discovery,
	}
}

// SelectEndpoint selects a healthy endpoint.
func (hlb *HealthAwareLoadBalancerService) SelectEndpoint(
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
