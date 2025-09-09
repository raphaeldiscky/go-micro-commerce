package config

import (
	"log"

	"github.com/spf13/viper"
)

// ClientConfig holds the configuration for external clients.
type ClientConfig struct {
	ProductGRPCHost     string `mapstructure:"PRODUCT_GRPC_HOST"`
	ProductGRPCPort     int    `mapstructure:"PRODUCT_GRPC_PORT"`
	FulfillmentGRPCPort int    `mapstructure:"FULFILLMENT_GRPC_PORT"`
	FulfillmentGRPCHost string `mapstructure:"FULFILLMENT_GRPC_HOST"`
	UseServiceDiscovery bool   `mapstructure:"USE_SERVICE_DISCOVERY"`
}

// initClientConfig initializes the client configuration.
func initClientConfig() *ClientConfig {
	viper.SetDefault("PRODUCT_GRPC_HOST", "0.0.0.0")
	viper.SetDefault("PRODUCT_GRPC_PORT", 50052)
	viper.SetDefault("FULFILLMENT_GRPC_HOST", "0.0.0.0")
	viper.SetDefault("FULFILLMENT_GRPC_PORT", 50055)
	viper.SetDefault("USE_SERVICE_DISCOVERY", true)

	clientCfg := &ClientConfig{}

	if err := viper.Unmarshal(&clientCfg); err != nil {
		log.Fatalf("error mapping client config: %v", err)
	}

	return clientCfg
}
