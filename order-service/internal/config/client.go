package config

import (
	"log"

	"github.com/spf13/viper"
)

// ClientConfig holds the configuration for external clients.
type ClientConfig struct {
	ProductGRPCHost     string `mapstructure:"PRODUCT_GRPC_HOST"`
	ProductGRPCPort     int    `mapstructure:"PRODUCT_GRPC_PORT"`
	UseServiceDiscovery bool   `mapstructure:"USE_SERVICE_DISCOVERY"`
}

// initClientConfig initializes the client configuration.
func initClientConfig() *ClientConfig {
	viper.SetDefault("PRODUCT_GRPC_HOST", "localhost")
	viper.SetDefault("PRODUCT_GRPC_PORT", 9502)
	viper.SetDefault("USE_SERVICE_DISCOVERY", true)

	clientCfg := &ClientConfig{}

	if err := viper.Unmarshal(&clientCfg); err != nil {
		log.Fatalf("error mapping client config: %v", err)
	}

	return clientCfg
}
