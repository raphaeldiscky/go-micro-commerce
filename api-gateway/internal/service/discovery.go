package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
)

// Discovery interface defines service discovery operations.
type Discovery interface {
	GetServiceEndpoint(serviceName string) (string, error)
	RegisterService(serviceName, address string, port int) error
	DeregisterService(serviceID string) error
	HealthCheck(serviceName string) bool
}

// ConsulDiscoveryService implements ServiceDiscovery using Consul.
type ConsulDiscoveryService struct {
	client *api.Client
	config config.ServiceDiscoveryConfig
	cache  map[string][]string
	mutex  sync.RWMutex
	logger logger.Logger
}

// NewConsulDiscoveryService creates a new Consul service discovery client.
func NewConsulDiscoveryService(
	cfg *config.ServiceDiscoveryConfig,
	appLogger logger.Logger,
) (*ConsulDiscoveryService, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Consul.Address
	consulConfig.Datacenter = cfg.Consul.Datacenter
	consulConfig.Token = cfg.Consul.Token

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	sd := &ConsulDiscoveryService{
		client: client,
		config: *cfg,
		cache:  make(map[string][]string),
		logger: appLogger,
	}

	// Start background cache refresh
	go sd.refreshCache()

	return sd, nil
}

// GetServiceEndpoint returns a healthy endpoint for the given service.
func (sd *ConsulDiscoveryService) GetServiceEndpoint(serviceName string) (string, error) {
	sd.mutex.RLock()
	endpoints, exists := sd.cache[serviceName]
	sd.mutex.RUnlock()

	if !exists || len(endpoints) == 0 {
		return "", fmt.Errorf("no healthy endpoints found for service: %s", serviceName)
	}

	// Simple round-robin for now (could be enhanced with load balancing)
	endpoint := endpoints[0]
	sd.logger.Debug("Selected endpoint", "service", serviceName, "endpoint", endpoint)

	return endpoint, nil
}

// RegisterService registers a service with Consul.
func (sd *ConsulDiscoveryService) RegisterService(serviceName, address string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", serviceName, address, port),
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
			Interval:                       "1s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return sd.client.Agent().ServiceRegister(registration)
}

// DeregisterService removes a service from Consul.
func (sd *ConsulDiscoveryService) DeregisterService(serviceID string) error {
	return sd.client.Agent().ServiceDeregister(serviceID)
}

// HealthCheck checks if a service is healthy.
func (sd *ConsulDiscoveryService) HealthCheck(serviceName string) bool {
	services, _, err := sd.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		sd.logger.Error("Health check failed", "service", serviceName, "error", err)

		return false
	}

	return len(services) > 0
}

// refreshCache periodically refreshes the service cache.
func (sd *ConsulDiscoveryService) refreshCache() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sd.updateCache()
	}
}

// updateCache updates the service endpoints cache.
func (sd *ConsulDiscoveryService) updateCache() {
	services, _, err := sd.client.Catalog().Services(nil)
	if err != nil {
		sd.logger.Error("Failed to fetch services", "error", err)

		return
	}

	newCache := make(map[string][]string)

	for serviceName := range services {
		healthyServices, _, err := sd.client.Health().Service(serviceName, "", true, nil)
		if err != nil {
			sd.logger.Error(
				"Failed to fetch healthy services",
				"service", serviceName,
				"error", err,
			)

			continue
		}

		var endpoints []string

		for _, service := range healthyServices {
			endpoint := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)
			endpoints = append(endpoints, endpoint)
		}

		if len(endpoints) > 0 {
			newCache[serviceName] = endpoints
		}
	}

	sd.mutex.Lock()
	sd.cache = newCache
	sd.mutex.Unlock()

	sd.logger.Debug("Service cache updated", "services", len(newCache))
}
