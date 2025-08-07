package config

import (
	"log"

	"github.com/spf13/viper"
)

// ServiceDiscoveryConfig holds service discovery configuration.
type ServiceDiscoveryConfig struct {
	Type             string `mapstructure:"SERVICE_DISCOVERY_TYPE"`
	Address          string `mapstructure:"SERVICE_DISCOVERY_ADDRESS"`
	ConsulAddress    string `mapstructure:"SERVICE_DISCOVERY_CONSUL_ADDRESS"`
	ConsulDatacenter string `mapstructure:"SERVICE_DISCOVERY_CONSUL_DATACENTER"`
	ConsulToken      string `mapstructure:"SERVICE_DISCOVERY_CONSUL_TOKEN"`
}

// initServiceDiscoveryConfig initializes the service discovery configuration from environment variables.
func initServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	// Set defaults
	viper.SetDefault("SERVICE_DISCOVERY_TYPE", "consul")
	viper.SetDefault("SERVICE_DISCOVERY_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_DATACENTER", "dc1")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_TOKEN", "")

	serviceDiscoveryConfig := &ServiceDiscoveryConfig{}

	if err := viper.Unmarshal(&serviceDiscoveryConfig); err != nil {
		log.Fatalf("error mapping service discovery config: %v", err)
	}

	return serviceDiscoveryConfig
}
