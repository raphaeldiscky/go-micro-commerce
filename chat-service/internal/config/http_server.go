package config

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// HTTPServerConfig holds the configuration for the HTTP server.
type HTTPServerConfig struct {
	Host                 string        `mapstructure:"HTTP_SERVER_HOST"`
	Port                 int           `mapstructure:"HTTP_SERVER_PORT"`
	GracePeriod          time.Duration `mapstructure:"HTTP_SERVER_GRACE_PERIOD"`
	RequestTimeoutPeriod time.Duration `mapstructure:"HTTP_SERVER_REQUEST_TIMEOUT_PERIOD"`
	ReadTimeout          time.Duration `mapstructure:"HTTP_SERVER_READ_TIMEOUT"`
	WriteTimeout         time.Duration `mapstructure:"HTTP_SERVER_WRITE_TIMEOUT"`
	IdleTimeout          time.Duration `mapstructure:"HTTP_SERVER_IDLE_TIMEOUT"`
	ReadHeaderTimeout    time.Duration `mapstructure:"HTTP_SERVER_READ_HEADER_TIMEOUT"`
	MaxHeaderBytes       int           `mapstructure:"HTTP_SERVER_MAX_HEADER_BYTES"`
	HSTSMaxAge           int           `mapstructure:"HTTP_SERVER_HSTS_MAX_AGE"`
	RateLimiter          rate.Limit    `mapstructure:"HTTP_SERVER_RATE_LIMITER"`
}

// initHTTPServerConfig initializes the HTTP server configuration from environment variables.
func initHTTPServerConfig() *HTTPServerConfig {
	// Set defaults
	viper.SetDefault("HTTP_SERVER_HOST", "localhost")
	viper.SetDefault("HTTP_SERVER_PORT", constant.HTTPServerPort)
	viper.SetDefault("HTTP_SERVER_GRACE_PERIOD", constant.HTTPServerGracePeriod)
	viper.SetDefault("HTTP_SERVER_REQUEST_TIMEOUT_PERIOD", constant.HTTPServerRequestTimeoutPeriod)
	viper.SetDefault("HTTP_SERVER_READ_TIMEOUT", constant.HTTPServerReadTimeout)
	viper.SetDefault("HTTP_SERVER_WRITE_TIMEOUT", constant.HTTPServerWriteTimeout)
	viper.SetDefault("HTTP_SERVER_IDLE_TIMEOUT", constant.HTTPServerIdleTimeout)
	viper.SetDefault("HTTP_SERVER_READ_HEADER_TIMEOUT", constant.HTTPServerReadHeaderTimeout)
	viper.SetDefault("HTTP_SERVER_MAX_HEADER_BYTES", constant.HTTPServerMaxHeaderBytes)
	viper.SetDefault("HTTP_SERVER_HSTS_MAX_AGE", constant.HTTPServerHSTSMaxAge)
	viper.SetDefault("HTTP_SERVER_RATE_LIMITER", constant.HTTPServerRateLimiter)

	httpServerConfig := &HTTPServerConfig{}
	if err := viper.Unmarshal(httpServerConfig); err != nil {
		panic(err)
	}

	return httpServerConfig
}
