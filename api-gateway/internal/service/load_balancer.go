// Package service provides various services for the API gateway, including circuit breakers, load balancing, and health checks.
package service

import (
	"errors"
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
