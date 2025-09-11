package config

import (
	"github.com/spf13/viper"
)

const (
	defaultClientProductGRPCPort     = 50052
	defaultClientFulfillmentGRPCPort = 50055
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
	viper.SetDefault("PRODUCT_GRPC_PORT", defaultClientProductGRPCPort)
	viper.SetDefault("FULFILLMENT_GRPC_HOST", "0.0.0.0")
	viper.SetDefault("FULFILLMENT_GRPC_PORT", defaultClientFulfillmentGRPCPort)
	viper.SetDefault("USE_SERVICE_DISCOVERY", true)

	clientCfg := &ClientConfig{}
	if err := viper.Unmarshal(&clientCfg); err != nil {
		panic(err)
	}

	return clientCfg
}
