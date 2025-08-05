// Package consul provides utilities for service registration and discovery with Consul.
package consul

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ServiceRegistration handles service registration with Consul.
type ServiceRegistration struct {
	client *api.Client
	logger *zap.Logger
}

// NewServiceRegistration creates a new service registration client.
func NewServiceRegistration(consulAddr string, logger *zap.Logger) (*ServiceRegistration, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &ServiceRegistration{
		client: client,
		logger: logger,
	}, nil
}

// ServiceConfig holds configuration for service registration.
type ServiceConfig struct {
	ServiceName    string
	ServiceID      string
	Address        string
	Port           int
	Tags           []string
	Meta           map[string]string
	HealthCheckURL string
	CheckInterval  string
	CheckTimeout   string
}

// Register registers a service with Consul.
func (s *ServiceRegistration) Register(config *ServiceConfig) error {
	// Generate service ID if not provided
	if config.ServiceID == "" {
		config.ServiceID = fmt.Sprintf("%s-%s", config.ServiceName, generateInstanceID())
	}

	// Set default health check URL if not provided
	if config.HealthCheckURL == "" {
		config.HealthCheckURL = fmt.Sprintf("http://%s:%d/health", config.Address, config.Port)
	}

	// Set default check intervals
	if config.CheckInterval == "" {
		config.CheckInterval = "30s"
	}

	if config.CheckTimeout == "" {
		config.CheckTimeout = "10s"
	}

	registration := &api.AgentServiceRegistration{
		ID:      config.ServiceID,
		Name:    config.ServiceName,
		Tags:    config.Tags,
		Address: config.Address,
		Port:    config.Port,
		Meta:    config.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           config.HealthCheckURL,
			Interval:                       config.CheckInterval,
			Timeout:                        config.CheckTimeout,
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	s.logger.Info("Registering service with Consul",
		zap.String("service_id", config.ServiceID),
		zap.String("service_name", config.ServiceName),
		zap.String("address", config.Address),
		zap.Int("port", config.Port),
		zap.Strings("tags", config.Tags),
	)

	if err := s.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	s.logger.Info("Service registered successfully",
		zap.String("service_id", config.ServiceID),
		zap.String("service_name", config.ServiceName),
	)

	return nil
}

// Deregister removes a service from Consul.
func (s *ServiceRegistration) Deregister(serviceID string) error {
	s.logger.Info("Deregistering service from Consul",
		zap.String("service_id", serviceID),
	)

	if err := s.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	s.logger.Info("Service deregistered successfully",
		zap.String("service_id", serviceID),
	)

	return nil
}

// GetHealthyServices returns healthy instances of a service.
func (s *ServiceRegistration) GetHealthyServices(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := s.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy services: %w", err)
	}

	return services, nil
}

// GetServiceAddress returns the address of a service.
func (s *ServiceRegistration) GetServiceAddress(serviceName string) (string, error) {
	services, err := s.GetHealthyServices(serviceName)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	service := services[0] // Simple load balancing - take first healthy instance

	return fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port), nil
}

// CreateTraefikTags creates Traefik-specific tags for service discovery.
func CreateTraefikTags(serviceName, host string, port int, additionalTags ...string) []string {
	tags := []string{
		"api",
		"v1",
		"traefik.enable=true",
		fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s`)", serviceName, host),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", serviceName, port),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints=web", serviceName),
	}

	tags = append(tags, additionalTags...)

	return tags
}

// generateInstanceID generates a unique instance ID for the service.
func generateInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get local IP if possible
	ip := getLocalIP()
	if ip != "" {
		hostname = ip
	}

	return hostname
}

// getLocalIP returns the local IP address of the machine.
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}

	if err := conn.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to close connection: %v\n", err)
	}

	localAddr := conn.LocalAddr()

	udpAddr, ok := localAddr.(*net.UDPAddr)
	if !ok {
		return ""
	}

	return udpAddr.IP.String()
}

// ParseConsulAddress parses consul address from environment variable.
func ParseConsulAddress() string {
	addr := os.Getenv("CONSUL_ADDRESS")
	if addr == "" {
		addr = "localhost:8500"
	}

	// Remove protocol if present
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")

	return addr
}

// GetServicePortFromEnv gets service port from environment variable.
func GetServicePortFromEnv(envVar, defaultPort string) (int, error) {
	portStr := os.Getenv(envVar)
	if portStr == "" {
		portStr = defaultPort
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port value: %s", portStr)
	}

	return port, nil
}
