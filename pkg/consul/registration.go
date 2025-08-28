// Package consul provides service registration with Consul for HTTP and gRPC services.
package consul

import (
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
)

// ServiceType represents the type of service (HTTP or gRPC).
type ServiceType string

const (
	// ServiceTypeHTTP represents HTTP service type for Consul registration.
	ServiceTypeHTTP ServiceType = "http"
	// ServiceTypeGRPC represents gRPC service type for Consul registration.
	ServiceTypeGRPC ServiceType = "grpc"
)

// ServiceRegistration handles Consul service registration.
type ServiceRegistration struct {
	client     *api.Client
	serviceIDs []string
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

// RegisterHTTP registers an HTTP service with Consul.
func (s *ServiceRegistration) RegisterHTTP(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeHTTP)
}

// RegisterGRPC registers a gRPC service with Consul.
func (s *ServiceRegistration) RegisterGRPC(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeGRPC)
}

// register registers a service with Consul based on service type.
func (s *ServiceRegistration) register(
	serviceName, address string,
	port int,
	serviceType ServiceType,
) error {
	// Generate unique service ID
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, hostname, port)

	var tags []string

	var check *api.AgentServiceCheck

	switch serviceType {
	case ServiceTypeHTTP:
		tags = []string{"http", "api", "microservice"}
		check = &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
			Interval:                       "30s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "60s",
		}
	case ServiceTypeGRPC:
		tags = []string{"grpc", "api", "microservice"}
		// For gRPC health checks, use TCP check since we have custom health method
		check = &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", address, port),
			Interval:                       "30s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "60s",
		}
	}

	// Service registration
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    tags,
		Check:   check,
		Meta: map[string]string{
			"protocol": string(serviceType),
		},
	}

	err = s.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service with consul: %w", err)
	}

	// Track registered service IDs
	s.serviceIDs = append(s.serviceIDs, serviceID)

	return nil
}

// Deregister removes all registered services from Consul.
func (s *ServiceRegistration) Deregister() error {
	if len(s.serviceIDs) == 0 {
		return nil
	}

	for _, serviceID := range s.serviceIDs {
		err := s.client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			return fmt.Errorf("failed to deregister service %s: %w", serviceID, err)
		}
	}

	return nil
}
