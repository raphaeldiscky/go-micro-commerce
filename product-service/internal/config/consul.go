package config

import (
	"log"

	"github.com/spf13/viper"
)

// ConsulConfig holds Consul service discovery configuration.
type ConsulConfig struct {
	Address     string `mapstructure:"CONSUL_ADDRESS"`
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServiceHost string `mapstructure:"SERVICE_HOST"`
	Enabled     bool   `mapstructure:"CONSUL_ENABLED"`
}

// initConsulConfig initializes the Consul configuration from environment variables.
func initConsulConfig() *ConsulConfig {
	consulConfig := &ConsulConfig{}

	if err := viper.Unmarshal(&consulConfig); err != nil {
		log.Fatalf("error mapping consul config: %v", err)
	}

	// Set defaults
	if consulConfig.Address == "" {
		consulConfig.Address = "localhost:8500"
	}

	if consulConfig.ServiceName == "" {
		consulConfig.ServiceName = "product-service"
	}

	if consulConfig.ServiceHost == "" {
		consulConfig.ServiceHost = "192.168.0.107" // Default host IP for Docker
	}

	// Enable by default if address is provided
	if !consulConfig.Enabled && consulConfig.Address != "" {
		consulConfig.Enabled = true
	}

	return consulConfig
}
