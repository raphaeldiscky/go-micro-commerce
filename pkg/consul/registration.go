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
	// ServiceTypeConnectRPC represents Connect-RPC service type for Consul registration.
	ServiceTypeConnectRPC ServiceType = "connectrpc"
	// ServiceTypeWebSocket represents WebSocket service type for Consul registration.
	ServiceTypeWebSocket ServiceType = "websocket"
)

// ServiceRegistration defines the interface for service registration.
type ServiceRegistration interface {
	RegisterHTTP(serviceName, address string, port int) error
	RegisterGRPC(serviceName, address string, port int) error
	RegisterConnectRPC(serviceName, address string, port int) error
	RegisterWebSocket(serviceName, address string, port int) error
	Deregister() error
}

// serviceRegistration handles Consul service registration.
type serviceRegistration struct {
	client     *api.Client
	logger     logger.Logger
	serviceIDs []string
}

// NewServiceRegistration creates a new Consul service registration client.
func NewServiceRegistration(
	consulAddr string,
	appLogger logger.Logger,
) (ServiceRegistration, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &serviceRegistration{
		client: client,
		logger: appLogger,
	}, nil
}

// GetServiceID generates a unique service ID based on service name, hostname, and port.
func (s *serviceRegistration) GetServiceID(serviceName, host string, port int) (string, error) {
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, host, port)

	return serviceID, nil
}

// RegisterHTTP registers an HTTP service with Consul.
func (s *serviceRegistration) RegisterHTTP(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeHTTP)
}

// RegisterGRPC registers a gRPC service with Consul.
func (s *serviceRegistration) RegisterGRPC(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeGRPC)
}

// RegisterConnectRPC registers a Connect-RPC service with Consul.
func (s *serviceRegistration) RegisterConnectRPC(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeConnectRPC)
}

// RegisterWebSocket registers a WebSocket service with Consul.
func (s *serviceRegistration) RegisterWebSocket(serviceName, address string, port int) error {
	return s.register(serviceName, address, port, ServiceTypeWebSocket)
}

// register registers a service with Consul based on service type.
func (s *serviceRegistration) register(
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
	case ServiceTypeConnectRPC:
		tags = []string{"connectrpc", "grpc", "http", "api", "microservice"}
		// Connect-RPC supports gRPC, HTTP, and gRPC-Web protocols over HTTP
		// Use HTTP health check since Connect-RPC serves over HTTP
		check = &api.AgentServiceCheck{
			HTTP: fmt.Sprintf(
				"http://%s/health",
				net.JoinHostPort(host, strconv.Itoa(port)),
			),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		}
	case ServiceTypeWebSocket:
		tags = []string{"websocket", "realtime", "microservice"}
		check = &api.AgentServiceCheck{
			HTTP: fmt.Sprintf(
				"http://%s/ws/health",
				net.JoinHostPort(host, strconv.Itoa(port)),
			),
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

// Deregister removes all registered services from Consul.
func (s *serviceRegistration) Deregister() error {
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
