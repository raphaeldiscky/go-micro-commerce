package config

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// SSEServerConfig holds the configuration for the SSE server.
type SSEServerConfig struct {
	Host        string        `mapstructure:"SSE_SERVER_HOST"`
	Port        int           `mapstructure:"SSE_SERVER_PORT"`
	Timeout     time.Duration `mapstructure:"SSE_SERVER_TIMEOUT"`
	RateLimiter rate.Limit    `mapstructure:"SSE_SERVER_RATE_LIMITER"`
}

// initSSEServerConfig initializes the SSE server configuration from environment variables.
func initSSEServerConfig() *SSEServerConfig {
	// Set defaults
	viper.SetDefault("SSE_SERVER_HOST", "localhost")
	viper.SetDefault("SSE_SERVER_PORT", constant.SSEServerPort)
	viper.SetDefault("SSE_SERVER_TIMEOUT", constant.SSEServerTimeout)
	viper.SetDefault("SSE_SERVER_RATE_LIMITER", constant.SSEServerRateLimiter)

	sseServerConfig := &SSEServerConfig{}
	if err := viper.Unmarshal(sseServerConfig); err != nil {
		panic(err)
	}

	return sseServerConfig
}
