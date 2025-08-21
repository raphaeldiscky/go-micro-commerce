// Package consul provides service registration with Consul for the services.
package consul

import (
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
)

// ServiceRegistration handles Consul service registration.
type ServiceRegistration struct {
	client    *api.Client
	serviceID string
}

// NewServiceRegistration creates a new Consul service registration client.
func NewServiceRegistration(consulAddr string) (*ServiceRegistration, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ServiceRegistration{
		client: client,
	}, nil
}

// Register registers the product service with Consul.
func (s *ServiceRegistration) Register(serviceName, address string, port int) error {
	// Generate unique service ID
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	s.serviceID = fmt.Sprintf("%s-%s-%d", serviceName, hostname, port)

	// Service registration
	registration := &api.AgentServiceRegistration{
		ID:      s.serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    []string{"http", "api", "microservice"},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
			Interval:                       "30s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	err = s.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service with consul: %w", err)
	}

	return nil
}

// Deregister removes the service from Consul.
func (s *ServiceRegistration) Deregister() error {
	if s.serviceID == "" {
		return nil
	}

	err := s.client.Agent().ServiceDeregister(s.serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	return nil
}
