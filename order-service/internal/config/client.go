package config

import (
	"log"

	"github.com/spf13/viper"
)

// ClientConfig holds the configuration for external clients.
type ClientConfig struct {
	ProductURL string `mapstructure:"PRODUCT_URL"`
}

// initClientConfig initializes the client configuration.
func initClientConfig() *ClientConfig {
	viper.SetDefault("PRODUCT_URL", "localhost:8081")

	clientCfg := &ClientConfig{}

	if err := viper.Unmarshal(&clientCfg); err != nil {
		log.Fatalf("error mapping client config: %v", err)
	}

	return clientCfg
}
