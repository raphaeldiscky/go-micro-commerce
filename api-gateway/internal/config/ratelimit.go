package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultRateLimitEnabled    = true
	defaultRateLimitRequests   = 100
	defaultRateLimitWindow     = 1 * time.Minute
	defaultRateLimitBurstLimit = 10
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
	viper.SetDefault("RATE_LIMIT_ENABLED", defaultRateLimitEnabled)
	viper.SetDefault("RATE_LIMIT_REQUESTS", defaultRateLimitRequests)
	viper.SetDefault("RATE_LIMIT_WINDOW", defaultRateLimitWindow)
	viper.SetDefault("RATE_LIMIT_BURST_LIMIT", defaultRateLimitBurstLimit)

	rateLimitConfig := &RateLimitConfig{}
	if err := viper.Unmarshal(rateLimitConfig); err != nil {
		panic(err)
	}

	return rateLimitConfig
}
