// Package config provides gRPC client configuration.
package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// GRPCClientConfig holds configuration for gRPC client creation.
type GRPCClientConfig struct {
	ServiceName           string        `mapstructure:"GRPC_SERVICE_NAME"`
	UseServiceDiscovery   bool          `mapstructure:"GRPC_USE_SERVICE_DISCOVERY"`
	ConsulAddress         string        `mapstructure:"GRPC_CONSUL_ADDRESS"`
	ConsulEnabled         bool          `mapstructure:"GRPC_CONSUL_ENABLED"`
	StaticAddress         string        `mapstructure:"GRPC_STATIC_ADDRESS"`
	StaticPort            int           `mapstructure:"GRPC_STATIC_PORT"`
	MaxAttempts           int64         `mapstructure:"GRPC_MAX_ATTEMPTS"`
	InitialBackoff        time.Duration `mapstructure:"GRPC_INITIAL_BACKOFF"`
	MaxBackoff            time.Duration `mapstructure:"GRPC_MAX_BACKOFF"`
	BackoffMultiplier     float64       `mapstructure:"GRPC_BACKOFF_MULTIPLIER"`
	RetryableStatusCodes  []string      `mapstructure:"GRPC_RETRYABLE_STATUS_CODES"`
	LoadBalancingPolicy   string        `mapstructure:"GRPC_LOAD_BALANCING_POLICY"`
	KeepaliveTime         time.Duration `mapstructure:"GRPC_KEEPALIVE_TIME"`
	KeepaliveTimeout      time.Duration `mapstructure:"GRPC_KEEPALIVE_TIMEOUT"`
	KeepalivePermitStream bool          `mapstructure:"GRPC_KEEPALIVE_PERMIT_STREAM"`
}

// initGRPCClientConfig creates a new gRPC client configuration from environment variables.
func initGRPCClientConfig(serviceName string) *GRPCClientConfig {
	grpcClientConfig := DefaultGRPCClientConfig(serviceName)

	if err := viper.Unmarshal(grpcClientConfig); err != nil {
		panic(err)
	}

	return grpcClientConfig
}

// DefaultGRPCClientConfig returns a default configuration for gRPC clients.
func DefaultGRPCClientConfig(serviceName string) *GRPCClientConfig {
	return &GRPCClientConfig{
		ServiceName:           serviceName,
		UseServiceDiscovery:   false,
		ConsulEnabled:         false,
		MaxAttempts:           constant.GRPCMaxAttempts,
		InitialBackoff:        constant.GRPCInitialBackoff,
		MaxBackoff:            constant.GRPCMaxBackoff,
		BackoffMultiplier:     constant.GRPCBackoffMultiplier,
		RetryableStatusCodes:  []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
		LoadBalancingPolicy:   "round_robin",
		KeepaliveTime:         constant.GRPCKeepaliveTime,
		KeepaliveTimeout:      constant.GRPCKeepaliveTimeout,
		KeepalivePermitStream: false,
	}
}

// SetStaticAddress sets the static address and port for the gRPC client.
func (c *GRPCClientConfig) SetStaticAddress(address string, port int) {
	c.StaticAddress = address
	c.StaticPort = port
	c.UseServiceDiscovery = false
}

// SetConsulDiscovery enables Consul service discovery with the given address.
func (c *GRPCClientConfig) SetConsulDiscovery(consulAddress string) {
	c.ConsulAddress = consulAddress
	c.ConsulEnabled = true
	c.UseServiceDiscovery = true
}
