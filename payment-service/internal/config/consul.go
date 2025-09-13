package config

import (
	"github.com/spf13/viper"
)

// ConsulConfig holds Consul service discovery configuration.
type ConsulConfig struct {
	Address string `mapstructure:"CONSUL_ADDRESS"`
	Enabled bool   `mapstructure:"CONSUL_ENABLED"`
}

// initConsulConfig initializes the Consul configuration from environment variables.
func initConsulConfig() *ConsulConfig {
	// Set defaults
	viper.SetDefault("CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("CONSUL_ENABLED", true)

	consulConfig := &ConsulConfig{}

	if err := viper.Unmarshal(consulConfig); err != nil {
		panic(err)
	}

	return consulConfig
}
