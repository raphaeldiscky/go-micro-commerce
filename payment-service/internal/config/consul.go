package config

import (
	"log/slog"

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
	// Set defaults
	viper.SetDefault("CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_NAME", "payment-service")
	viper.SetDefault("SERVICE_HOST", "192.168.0.107")
	viper.SetDefault("CONSUL_ENABLED", true)

	consulConfig := &ConsulConfig{}

	if err := viper.Unmarshal(&consulConfig); err != nil {
		slog.Error("error mapping consul config", "err", err)
	}

	return consulConfig
}
