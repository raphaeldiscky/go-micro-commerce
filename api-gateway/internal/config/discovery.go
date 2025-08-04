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
	serviceDiscoveryConfig := &ServiceDiscoveryConfig{}

	if err := viper.Unmarshal(&serviceDiscoveryConfig); err != nil {
		log.Fatalf("error mapping service discovery config: %v", err)
	}

	return serviceDiscoveryConfig
}
