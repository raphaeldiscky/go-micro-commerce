package config

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

const (
	defaultHTTPServerPort                            = 8087
	defaultHTTPServerGracePeriod                     = 10 * time.Second
	defaultHTTPServerRequestTimeoutPeriod            = 30 * time.Second
	defaultHTTPServerReadTimeout                     = 30 * time.Second
	defaultHTTPServerWriteTimeout                    = 30 * time.Second
	defaultHTTPServerIdleTimeout                     = 120 * time.Second
	defaultHTTPServerReadHeaderTimeout               = 10 * time.Second
	defaultHTTPServerMaxHeaderBytes                  = 1 << 20
	defaultHTTPServerHSTSMaxAge                      = 3600
	defaultHTTPServerRateLimiter          rate.Limit = 1000
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
	viper.SetDefault("HTTP_SERVER_PORT", defaultHTTPServerPort)
	viper.SetDefault("HTTP_SERVER_GRACE_PERIOD", defaultHTTPServerGracePeriod)
	viper.SetDefault("HTTP_SERVER_REQUEST_TIMEOUT_PERIOD", defaultHTTPServerRequestTimeoutPeriod)
	viper.SetDefault("HTTP_SERVER_READ_TIMEOUT", defaultHTTPServerReadTimeout)
	viper.SetDefault("HTTP_SERVER_WRITE_TIMEOUT", defaultHTTPServerWriteTimeout)
	viper.SetDefault("HTTP_SERVER_IDLE_TIMEOUT", defaultHTTPServerIdleTimeout)
	viper.SetDefault("HTTP_SERVER_READ_HEADER_TIMEOUT", defaultHTTPServerReadHeaderTimeout)
	viper.SetDefault("HTTP_SERVER_MAX_HEADER_BYTES", defaultHTTPServerMaxHeaderBytes)
	viper.SetDefault("HTTP_SERVER_HSTS_MAX_AGE", defaultHTTPServerHSTSMaxAge)
	viper.SetDefault("HTTP_SERVER_RATE_LIMITER", defaultHTTPServerRateLimiter)

	httpServerConfig := &HTTPServerConfig{}

	return httpServerConfig
}
