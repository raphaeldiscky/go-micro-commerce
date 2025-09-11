package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultServiceDiscoveryTimeout = 5 * time.Second
	defaultConsulRefreshInterval   = 5 * time.Second
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
	Address         string        `mapstructure:"CONSUL_ADDRESS"`
	Token           string        `mapstructure:"CONSUL_TOKEN"`
	Datacenter      string        `mapstructure:"CONSUL_DATACENTER"`
	RefreshInterval time.Duration `mapstructure:"CONSUL_REFRESH_INTERVAL"`
}

// initServiceDiscoveryConfig initializes the service discovery configuration from environment variables.
func initServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	viper.SetDefault("SERVICE_DISCOVERY_TYPE", "consul")
	viper.SetDefault("SERVICE_DISCOVERY_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_TIMEOUT", defaultServiceDiscoveryTimeout)

	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_TOKEN", "")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_DATACENTER", "dc1")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_REFRESH_INTERVAL", defaultConsulRefreshInterval)

	serviceDiscoveryConfig := &ServiceDiscoveryConfig{}
	if err := viper.Unmarshal(serviceDiscoveryConfig); err != nil {
		panic(err)
	}

	return serviceDiscoveryConfig
}
