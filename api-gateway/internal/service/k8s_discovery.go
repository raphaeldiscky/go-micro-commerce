package service

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
)

// KubernetesDiscoveryService implements ServiceDiscovery using Kubernetes DNS.
type KubernetesDiscoveryService struct {
	logger    logger.Logger
	namespace string
	prefix    string            // K8s service name prefix (e.g., "local-", "dev-")
	endpoints map[string]string // service name -> endpoint mapping
}

// NewKubernetesDiscoveryService creates a new Kubernetes DNS service discovery client.
func NewKubernetesDiscoveryService(
	cfg *config.ServiceDiscoveryConfig,
	appLogger logger.Logger,
) *KubernetesDiscoveryService {
	sd := &KubernetesDiscoveryService{
		logger:    appLogger,
		namespace: cfg.K8sNamespace,
		prefix:    cfg.K8sServicePrefix,
		endpoints: make(map[string]string),
	}

	// Pre-populate service endpoints for known services
	sd.initializeServiceEndpoints()

	return sd
}

const (
	// HTTP server ports.
	authServicePort         = 8081
	productServicePort      = 8082
	orderServicePort        = 8083
	paymentServicePort      = 8084
	fulfillmentServicePort  = 8085
	notificationServicePort = 8086
	searchServicePort       = 8087
	chatServicePort         = 8088
	cartServicePort         = 8089
	graphQLGatewayPort      = 80 // Apollo Router K8s service port (maps to container port 4000)

	// Specialized server ports (gRPC/connect-RPC, SSE, WebSocket).
	productServiceGRPCPort     = 50052
	notificationServiceSSEPort = 9086
	chatServiceWSPort          = 9098
)

// initializeServiceEndpoints initializes the service endpoints based on Kubernetes DNS convention.
func (sd *KubernetesDiscoveryService) initializeServiceEndpoints() {
	// Kubernetes DNS convention: <service-name>.<namespace>.svc.cluster.local:<port>
	// For services in the same namespace, we can use short form: <service-name>:<port>
	services := map[string]int{
		// HTTP server ports
		"auth-service":         authServicePort,
		"product-service":      productServicePort,
		"order-service":        orderServicePort,
		"payment-service":      paymentServicePort,
		"fulfillment-service":  fulfillmentServicePort,
		"notification-service": notificationServicePort,
		"search-service":       searchServicePort,
		"chat-service":         chatServicePort,
		"cart-service":         cartServicePort,
		"apollo-router":        graphQLGatewayPort, // K8s service name for GraphQL Gateway

		// Specialized servers (same K8s service, different ports)
		"product-service-grpc":     productServiceGRPCPort,     // Connect-RPC server
		"notification-service-sse": notificationServiceSSEPort, // SSE subscription server
		"chat-service-ws":          chatServiceWSPort,          // WebSocket server
	}

	// Map logical service names to actual K8s service names
	// This handles cases where multiple ports are exposed on the same service
	// The prefix (e.g., "local-") is applied dynamically based on configuration
	serviceNameMapping := map[string]string{
		"product-service-grpc":     sd.prefix + "product-service",
		"notification-service-sse": sd.prefix + "notification-service",
		"chat-service-ws":          sd.prefix + "chat-service",
	}

	for serviceName, port := range services {
		// Resolve actual K8s service name
		// First, check if there's a custom mapping (for specialized services)
		// Otherwise, apply the prefix to the base service name
		var actualServiceName string
		if mappedName, exists := serviceNameMapping[serviceName]; exists {
			actualServiceName = mappedName
		} else {
			// Apply prefix to regular service names
			actualServiceName = sd.prefix + serviceName
		}

		var endpoint string
		if sd.namespace != "" {
			// Use FQDN if namespace is specified
			endpoint = fmt.Sprintf(
				"http://%s.%s.svc.cluster.local:%d",
				actualServiceName,
				sd.namespace,
				port,
			)
		} else {
			// Use short form (assumes same namespace)
			endpoint = net.JoinHostPort(actualServiceName, strconv.Itoa(port))
		}

		sd.endpoints[serviceName] = endpoint
		sd.logger.Debugf("Registered K8s service endpoint: %s -> %s", serviceName, endpoint)
	}
}

// GetServiceEndpoint returns the endpoint for the given service using Kubernetes DNS.
func (sd *KubernetesDiscoveryService) GetServiceEndpoint(serviceName string) (string, error) {
	endpoint, exists := sd.endpoints[serviceName]
	if !exists {
		return "", fmt.Errorf("no endpoint configured for service: %s", serviceName)
	}

	sd.logger.Debugf("Resolved K8s service endpoint for %s: %s", serviceName, endpoint)

	return endpoint, nil
}

// RegisterService is a no-op for Kubernetes as services are registered via K8s Service resources.
func (sd *KubernetesDiscoveryService) RegisterService(serviceName, _ string, _ int) error {
	sd.logger.Infof(
		"Kubernetes service registration is managed by K8s Service resources, skipping registration for %s",
		serviceName,
	)

	return nil
}

// DeregisterService is a no-op for Kubernetes.
func (sd *KubernetesDiscoveryService) DeregisterService(serviceID string) error {
	sd.logger.Infof(
		"Kubernetes service deregistration is managed by K8s, skipping deregistration for %s",
		serviceID,
	)

	return nil
}

// HealthCheck checks if a service is healthy by attempting a connection.
func (sd *KubernetesDiscoveryService) HealthCheck(serviceName string) bool {
	endpoint, exists := sd.endpoints[serviceName]
	if !exists {
		sd.logger.Error("No endpoint found for service", "service", serviceName)
		return false
	}

	// Parse the endpoint to get host:port
	// endpoint format: http://service-name:port or http://service-name.namespace.svc.cluster.local:port
	// We need to extract host:port for TCP check
	host, portStr := sd.parseEndpoint(endpoint)
	if host == "" || portStr == "" {
		sd.logger.Error("Failed to parse endpoint", "service", serviceName, "endpoint", endpoint)
		return false
	}

	// Attempt TCP connection
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort(host, portStr))
	if err != nil {
		sd.logger.Error("Health check failed", "service", serviceName, "error", err)
		return false
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			sd.logger.Warn("Failed to close health check connection", "error", closeErr)
		}
	}()

	return true
}

// parseEndpoint extracts host and port from endpoint URL.
func (sd *KubernetesDiscoveryService) parseEndpoint(endpoint string) (string, string) {
	// Remove http:// or https:// prefix
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
	} else if len(endpoint) > 8 && endpoint[:8] == "https://" {
		endpoint = endpoint[8:]
	}

	// Split host:port
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		sd.logger.Warn("Failed to split host:port", "endpoint", endpoint, "error", err)
		return "", ""
	}

	// Validate port
	if _, err = strconv.Atoi(portStr); err != nil {
		sd.logger.Warn("Invalid port number", "port", portStr, "error", err)
		return "", ""
	}

	return host, portStr
}
