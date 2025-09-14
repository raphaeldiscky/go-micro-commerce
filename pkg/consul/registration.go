// Package consul provides service registration with Consul for HTTP and gRPC services.
package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// ServiceType represents the type of service (HTTP or gRPC).
type ServiceType string

const (
	// ServiceTypeHTTP represents HTTP service type for Consul registration.
	ServiceTypeHTTP ServiceType = "http"
	// ServiceTypeGRPC represents gRPC service type for Consul registration.
	ServiceTypeGRPC ServiceType = "grpc"
)

// ServiceRegistrationInterface defines the interface for service registration.
type ServiceRegistrationInterface interface {
	RegisterHTTP(serviceName, address string, port int) error
	RegisterGRPC(serviceName, address string, port int) error
	DeregisterGRPC(serviceID string) error
	DeregisterHTTP(serviceID string) error
}

// ServiceRegistration handles Consul service registration.
type ServiceRegistration struct {
	client     *api.Client
	logger     logger.Logger
	serviceIDs []string
}

// NewServiceRegistration creates a new Consul service registration client.
func NewServiceRegistration(
	consulAddr string,
	appLogger logger.Logger,
) (*ServiceRegistration, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ServiceRegistration{
		client: client,
		logger: appLogger,
	}, nil
}

// GetServiceID generates a unique service ID based on service name, hostname, and port.
func (s *ServiceRegistration) GetServiceID(serviceName, host string, port int) (string, error) {
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, host, port)

	return serviceID, nil
}

// RegisterHTTP registers an HTTP service with Consul.
func (s *ServiceRegistration) RegisterHTTP(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeHTTP)
}

// DeregisterHTTP deregisters an HTTP service from Consul.
func (s *ServiceRegistration) DeregisterHTTP(serviceID string) error {
	return s.deregister(serviceID, ServiceTypeHTTP)
}

// RegisterGRPC registers a gRPC service with Consul.
func (s *ServiceRegistration) RegisterGRPC(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeGRPC)
}

// DeregisterGRPC deregisters a gRPC service from Consul.
func (s *ServiceRegistration) DeregisterGRPC(serviceID string) error {
	return s.deregister(serviceID, ServiceTypeGRPC)
}

// register registers a service with Consul based on service type.
func (s *ServiceRegistration) register(
	serviceName, host string,
	port int,
	serviceType ServiceType,
) error {
	serviceID, err := s.GetServiceID(serviceName, host, port)
	if err != nil {
		return fmt.Errorf("failed to get service ID: %w", err)
	}

	var tags []string

	var check *api.AgentServiceCheck

	switch serviceType {
	case ServiceTypeHTTP:
		tags = []string{"http", "api", "microservice"}
		check = &api.AgentServiceCheck{
			HTTP: fmt.Sprintf(
				"http://%s/health",
				net.JoinHostPort(host, strconv.Itoa(port)),
			),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		}
	case ServiceTypeGRPC:
		tags = []string{"grpc", "api", "microservice"}
		// For gRPC health checks, use TCP check since we have custom health method
		check = &api.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", host, port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		}
	}

	// Service registration
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: host,
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

	s.serviceIDs = append(s.serviceIDs, serviceID)

	s.logger.Infof("Service %s registered with Consul", serviceID)

	return nil
}

// deregister removes a service from Consul by service type.
func (s *ServiceRegistration) deregister(serviceID string, serviceType ServiceType) error {
	err := s.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister %s service %s: %w", serviceType, serviceID, err)
	}

	s.logger.Infof("%s service %s deregistered from Consul", serviceType, serviceID)

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
			// Log the error but don't fail the entire deregistration process
			s.logger.Errorf("Failed to deregister service %s: %v", serviceID, err)
			continue
		}

		s.logger.Infof("Service %s deregistered from Consul", serviceID)
	}

	s.logger.Infof("Deregistration process completed")

	return nil
}
