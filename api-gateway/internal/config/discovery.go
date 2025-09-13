package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/constant"
)

// ServiceDiscoveryConfig holds service discovery configuration.
type ServiceDiscoveryConfig struct {
	Type                  string        `mapstructure:"SERVICE_DISCOVERY_TYPE"`
	Address               string        `mapstructure:"SERVICE_DISCOVERY_ADDRESS"`
	Timeout               time.Duration `mapstructure:"SERVICE_DISCOVERY_TIMEOUT"`
	ConsulAddress         string        `mapstructure:"SERVICE_DISCOVERY_CONSUL"`
	ConsulToken           string        `mapstructure:"SERVICE_DISCOVERY_CONSUL_TOKEN"`
	ConsulDatacenter      string        `mapstructure:"SERVICE_DISCOVERY_CONSUL_DATACENTER"`
	ConsulRefreshInterval time.Duration `mapstructure:"SERVICE_DISCOVERY_CONSUL_REFRESH_INTERVAL"`
}

// initServiceDiscoveryConfig initializes the service discovery configuration from environment variables.
func initServiceDiscoveryConfig() *ServiceDiscoveryConfig {
	viper.SetDefault("SERVICE_DISCOVERY_TYPE", "consul")
	viper.SetDefault("SERVICE_DISCOVERY_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_TIMEOUT", constant.ServiceDiscoveryTimeout)

	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_TOKEN", "token")
	viper.SetDefault("SERVICE_DISCOVERY_CONSUL_DATACENTER", "dc1")
	viper.SetDefault(
		"SERVICE_DISCOVERY_CONSUL_REFRESH_INTERVAL",
		constant.ServiceDiscoveryConsulRefreshInterval,
	)

	serviceDiscoveryConfig := &ServiceDiscoveryConfig{}
	if err := viper.Unmarshal(serviceDiscoveryConfig); err != nil {
		panic(err)
	}

	return serviceDiscoveryConfig
}
