package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// ServiceDiscoveryConfig holds service discovery configuration.
type ServiceDiscoveryConfig struct {
	Type    string        `mapstructure:"SERVICE_DISCOVERY_TYPE"`
	Address string        `mapstructure:"SERVICE_DISCOVERY_ADDRESS"`
	Timeout time.Duration `mapstructure:"SERVICE_DISCOVERY_TIMEOUT"`
	Consul  ConsulConfig  `mapstructure:"SERVICE_DISCOVERY_CONSUL"`
}

// ConsulConfig holds Consul-specific configuration.
type ConsulConfig struct {
	Address    string `mapstructure:"CONSUL_ADDRESS"`
	Token      string `mapstructure:"CONSUL_TOKEN"`
	Datacenter string `mapstructure:"CONSUL_DATACENTER"`
}

// initServiceDiscoveryConfig initializes the service discovery configuration from environment variables.
func initServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	viper.SetDefault("SERVICE_DISCOVERY_TYPE", "consul")
	viper.SetDefault("SERVICE_DISCOVERY_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_TIMEOUT", 5*time.Second)

	serviceDiscoveryConfig := &ServiceDiscoveryConfig{}

	if err := viper.Unmarshal(&serviceDiscoveryConfig); err != nil {
		log.Fatalf("error mapping service discovery config: %v", err)
	}

	return serviceDiscoveryConfig
}
