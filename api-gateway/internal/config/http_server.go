package config

import (
	"log/slog"

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
	viper.SetDefault("HTTP_SERVER_HOST", "localhost")
	viper.SetDefault("HTTP_SERVER_PORT", 8080)
	viper.SetDefault("HTTP_SERVER_GRACE_PERIOD", 5)
	viper.SetDefault("HTTP_SERVER_REQUEST_TIMEOUT_PERIOD", 10)

	httpServerConfig := &HTTPServerConfig{}

	if err := viper.Unmarshal(&httpServerConfig); err != nil {
		slog.Error("error mapping http server config", "err", err)
	}

	return httpServerConfig
}
