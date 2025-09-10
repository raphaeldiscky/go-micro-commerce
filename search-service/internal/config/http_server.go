package config

import (
	"log"

	"github.com/spf13/viper"
)

// HTTPServerConfig holds the configuration for the HTTP server.
type HTTPServerConfig struct {
	Host                 string `mapstructure:"HTTP_SERVER_HOST"`
	Port                 int    `mapstructure:"HTTP_SERVER_PORT"`
	GracePeriod          int    `mapstructure:"HTTP_SERVER_GRACE_PERIOD"`
	RequestTimeoutPeriod int    `mapstructure:"HTTP_SERVER_REQUEST_TIMEOUT_PERIOD"`
}

// initHTTPServerConfig initializes the HTTP server configuration from environment variables.
func initHTTPServerConfig() *HTTPServerConfig {
	// Set defaults
	viper.SetDefault("HTTP_SERVER_HOST", "localhost")
	viper.SetDefault("HTTP_SERVER_PORT", 8087)
	viper.SetDefault("HTTP_SERVER_GRACE_PERIOD", 10)
	viper.SetDefault("HTTP_SERVER_REQUEST_TIMEOUT_PERIOD", 30)

	httpServerConfig := &HTTPServerConfig{}

	if err := viper.Unmarshal(&httpServerConfig); err != nil {
		log.Fatalf("error mapping http server config: %v", err)
	}

	return httpServerConfig
}
