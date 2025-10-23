package config

import (
	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// ClientConfig holds the configuration for external clients.
type ClientConfig struct {
	ProductGRPCHost     string `mapstructure:"CLIENT_PRODUCT_GRPC_HOST"`
	ProductGRPCPort     int    `mapstructure:"CLIENT_PRODUCT_GRPC_PORT"`
	FulfillmentGRPCPort int    `mapstructure:"CLIENT_FULFILLMENT_GRPC_PORT"`
	FulfillmentGRPCHost string `mapstructure:"CLIENT_FULFILLMENT_GRPC_HOST"`
	UseServiceDiscovery bool   `mapstructure:"CLIENT_USE_SERVICE_DISCOVERY"`
}

// initClientConfig initializes the client configuration.
func initClientConfig() *ClientConfig {
	viper.SetDefault("CLIENT_PRODUCT_GRPC_HOST", "172.19.0.1")
	viper.SetDefault("CLIENT_PRODUCT_GRPC_PORT", constant.ClientProductGRPCPort)
	viper.SetDefault("CLIENT_FULFILLMENT_GRPC_HOST", "172.19.0.1")
	viper.SetDefault("CLIENT_FULFILLMENT_GRPC_PORT", constant.ClientFulfillmentGRPCPort)
	viper.SetDefault("CLIENT_USE_SERVICE_DISCOVERY", true)

	clientCfg := &ClientConfig{}
	if err := viper.Unmarshal(&clientCfg); err != nil {
		panic(err)
	}

	return clientCfg
}
