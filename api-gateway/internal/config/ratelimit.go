package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
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
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", 1*time.Minute)
	viper.SetDefault("RATE_LIMIT_BURST_LIMIT", 10)

	rateLimitConfig := &RateLimitConfig{}

	if err := viper.Unmarshal(&rateLimitConfig); err != nil {
		log.Fatalf("error mapping rate limit config: %v", err)
	}

	return rateLimitConfig
}
