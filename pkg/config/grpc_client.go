// Package config provides gRPC client configuration.
package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultGRPCMaxAttempts      = 3
	defaultGRPCMultiplier       = 2.0
	defaultGRPCKeepaliveTime    = 30 * time.Second
	defaultGRPCKeepaliveTimeout = 5 * time.Second
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
	InitialBackoff        string        `mapstructure:"GRPC_INITIAL_BACKOFF"`
	MaxBackoff            string        `mapstructure:"GRPC_MAX_BACKOFF"`
	BackoffMultiplier     float64       `mapstructure:"GRPC_BACKOFF_MULTIPLIER"`
	RetryableStatusCodes  []string      `mapstructure:"GRPC_RETRYABLE_STATUS_CODES"`
	LoadBalancingPolicy   string        `mapstructure:"GRPC_LOAD_BALANCING_POLICY"`
	KeepaliveTime         time.Duration `mapstructure:"GRPC_KEEPALIVE_TIME"`
	KeepaliveTimeout      time.Duration `mapstructure:"GRPC_KEEPALIVE_TIMEOUT"`
	KeepalivePermitStream bool          `mapstructure:"GRPC_KEEPALIVE_PERMIT_STREAM"`
}

// DefaultGRPCClientConfig returns a default configuration for gRPC clients.
func DefaultGRPCClientConfig(serviceName string) *GRPCClientConfig {
	return &GRPCClientConfig{
		ServiceName:           serviceName,
		UseServiceDiscovery:   false,
		ConsulEnabled:         false,
		MaxAttempts:           defaultGRPCMaxAttempts,
		InitialBackoff:        "0.1s",
		MaxBackoff:            "1s",
		BackoffMultiplier:     defaultGRPCMultiplier,
		RetryableStatusCodes:  []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
		LoadBalancingPolicy:   "round_robin",
		KeepaliveTime:         defaultGRPCKeepaliveTime,
		KeepaliveTimeout:      defaultGRPCKeepaliveTimeout,
		KeepalivePermitStream: false,
	}
}

// initGRPCClientConfig creates a new gRPC client configuration from environment variables.
func initGRPCClientConfig(serviceName string) *GRPCClientConfig {
	config := DefaultGRPCClientConfig(serviceName)

	// Override with environment variables if they exist
	if viper.IsSet("GRPC_USE_SERVICE_DISCOVERY") {
		config.UseServiceDiscovery = viper.GetBool("GRPC_USE_SERVICE_DISCOVERY")
	}

	if viper.IsSet("GRPC_CONSUL_ADDRESS") {
		config.ConsulAddress = viper.GetString("GRPC_CONSUL_ADDRESS")
	}

	if viper.IsSet("GRPC_CONSUL_ENABLED") {
		config.ConsulEnabled = viper.GetBool("GRPC_CONSUL_ENABLED")
	}

	if viper.IsSet("GRPC_MAX_ATTEMPTS") {
		config.MaxAttempts = viper.GetInt64("GRPC_MAX_ATTEMPTS")
	}

	if viper.IsSet("GRPC_INITIAL_BACKOFF") {
		config.InitialBackoff = viper.GetString("GRPC_INITIAL_BACKOFF")
	}

	if viper.IsSet("GRPC_MAX_BACKOFF") {
		config.MaxBackoff = viper.GetString("GRPC_MAX_BACKOFF")
	}

	if viper.IsSet("GRPC_BACKOFF_MULTIPLIER") {
		config.BackoffMultiplier = viper.GetFloat64("GRPC_BACKOFF_MULTIPLIER")
	}

	if viper.IsSet("GRPC_LOAD_BALANCING_POLICY") {
		config.LoadBalancingPolicy = viper.GetString("GRPC_LOAD_BALANCING_POLICY")
	}

	if viper.IsSet("GRPC_KEEPALIVE_TIME") {
		config.KeepaliveTime = viper.GetDuration("GRPC_KEEPALIVE_TIME")
	}

	if viper.IsSet("GRPC_KEEPALIVE_TIMEOUT") {
		config.KeepaliveTimeout = viper.GetDuration("GRPC_KEEPALIVE_TIMEOUT")
	}

	if viper.IsSet("GRPC_KEEPALIVE_PERMIT_STREAM") {
		config.KeepalivePermitStream = viper.GetBool("GRPC_KEEPALIVE_PERMIT_STREAM")
	}

	return config
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
