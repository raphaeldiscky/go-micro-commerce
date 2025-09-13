package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/constant"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled    bool          `mapstructure:"RATE_LIMIT_ENABLED"`
	Requests   int           `mapstructure:"RATE_LIMIT_REQUESTS"`
	Window     time.Duration `mapstructure:"RATE_LIMIT_WINDOW"`
	BurstLimit int           `mapstructure:"RATE_LIMIT_BURST_LIMIT"`
}

// initRateLimitConfig initializes the rate limit configuration from environment variables.
func initRateLimitConfig() *RateLimitConfig {
	viper.SetDefault("RATE_LIMIT_ENABLED", constant.RateLimitEnabled)
	viper.SetDefault("RATE_LIMIT_REQUESTS", constant.RateLimitRequests)
	viper.SetDefault("RATE_LIMIT_WINDOW", constant.RateLimitWindow)
	viper.SetDefault("RATE_LIMIT_BURST_LIMIT", constant.RateLimitBurstLimit)

	rateLimitConfig := &RateLimitConfig{}
	if err := viper.Unmarshal(rateLimitConfig); err != nil {
		panic(err)
	}

	return rateLimitConfig
}
