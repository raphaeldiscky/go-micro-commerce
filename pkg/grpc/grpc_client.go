// Package grpc provides reusable gRPC client utilities with Consul service discovery support.
package grpc

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// Client provides a reusable gRPC client with Consul service discovery support.
type Client struct {
	conn         *grpc.ClientConn
	consulClient *api.Client
	config       *config.GRPCClientConfig
}

// NewGRPCClient creates a new GRPCClient instance with the specified configuration.
func NewGRPCClient(cfg *config.GRPCClientConfig) (*Client, error) {
	var conn *grpc.ClientConn

	var consulClient *api.Client

	var err error

	if shouldUseServiceDiscovery(cfg) {
		conn, consulClient, err = createConsulConnection(cfg)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = createStaticConnection(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		conn:         conn,
		consulClient: consulClient,
		config:       cfg,
	}, nil
}

// GetConnection returns the underlying gRPC connection.
func (gc *Client) GetConnection() *grpc.ClientConn {
	return gc.conn
}

// Close closes the gRPC connection.
func (gc *Client) Close() error {
	return gc.conn.Close()
}

// shouldUseServiceDiscovery checks if service discovery should be used.
func shouldUseServiceDiscovery(cfg *config.GRPCClientConfig) bool {
	return cfg.UseServiceDiscovery && cfg.ConsulEnabled
}

// createConsulConnection creates a gRPC connection using Consul service discovery.
func createConsulConnection(
	cfg *config.GRPCClientConfig,
) (*grpc.ClientConn, *api.Client, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.ConsulAddress

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	address, err := getServiceAddress(consulClient, cfg.ServiceName)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to get %s address from consul: %w",
			cfg.ServiceName,
			err,
		)
	}

	conn, err := createGRPCConnection(address, cfg)
	if err != nil {
		return nil, nil, err
	}

	return conn, consulClient, nil
}

// createStaticConnection creates a gRPC connection using static configuration.
func createStaticConnection(cfg *config.GRPCClientConfig) (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%d", cfg.StaticAddress, cfg.StaticPort)

	return createGRPCConnection(address, cfg)
}

// createGRPCConnection creates a gRPC connection with common options and resilience features.
func createGRPCConnection(
	address string,
	cfg *config.GRPCClientConfig,
) (*grpc.ClientConn, error) {
	clientAuth := NewClientAuthInterceptor()

	// Configure keepalive parameters for automatic reconnection
	kacp := keepalive.ClientParameters{
		Time:                cfg.KeepaliveTime,
		Timeout:             cfg.KeepaliveTimeout,
		PermitWithoutStream: cfg.KeepalivePermitStream,
	}

	// Build service config for retry policy
	serviceConfig := fmt.Sprintf(`{
		"methodConfig": [{
			"name": [{"service": "%s"}],
			"retryPolicy": {
				"MaxAttempts": %d,
				"InitialBackoff": "%.3fs",
				"MaxBackoff": "%.3fs",
				"BackoffMultiplier": %f,
				"RetryableStatusCodes": [%s]
			}
		}],
		"loadBalancingPolicy": "%s"
	}`, getServiceNameFromFullService(cfg.ServiceName), cfg.MaxAttempts,
		cfg.InitialBackoff.Seconds(), cfg.MaxBackoff.Seconds(), cfg.BackoffMultiplier,
		buildRetryableCodesString(cfg.RetryableStatusCodes), cfg.LoadBalancingPolicy)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(clientAuth.ForwardUserAuth()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultServiceConfig(serviceConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.ServiceName, err)
	}

	return conn, nil
}

// getServiceAddress retrieves the service address from Consul.
func getServiceAddress(consulClient *api.Client, serviceName string) (string, error) {
	services, _, err := consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query consul for service %s: %w", serviceName, err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	// Use the first healthy instance
	service := services[0].Service

	return fmt.Sprintf("%s:%d", service.Address, service.Port), nil
}

// getServiceNameFromFullService extracts the service name from the full service name for protobuf.
// For example, "product-service-grpc" becomes "product.ProductService".
func getServiceNameFromFullService(serviceName string) string {
	switch serviceName {
	case constant.GRPCServiceNameProduct:
		return "product.ProductService"
	case constant.GRPCServiceNameOrder:
		return "order.OrderService"
	case constant.GRPCServiceNamePayment:
		return "payment.PaymentService"
	case constant.GRPCServiceNameNotification:
		return "notification.NotificationService"
	case constant.GRPCServiceNameAuth:
		return "auth.AuthService"
	default:
		// Fallback: try to convert service name to protobuf service name
		return serviceName
	}
}

// buildRetryableCodesString builds a JSON array string from retryable status codes.
func buildRetryableCodesString(codes []string) string {
	if len(codes) == 0 {
		return ""
	}

	result := ""

	for i, code := range codes {
		if i > 0 {
			result += ", "
		}

		result += fmt.Sprintf(`%q`, code)
	}

	return result
}
