// Package config provides configuration management for the API gateway.
package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App              *AppConfig
	JWT              *JWTConfig
	HTTPServer       *HTTPServerConfig
	ServiceDiscovery *ServiceDiscoveryConfig
	RateLimit        *RateLimitConfig
	Tracing          *TracingConfig
	Metrics          *MetricsConfig
	CircuitBreaker   *CircuitBreakerConfig
}

// LoadConfig loads configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	//nolint:errcheck // .env file not required when using environment variables
	_ = viper.ReadInConfig()

	cfg := &Config{
		App:              initAppConfig(),
		JWT:              initJWTConfig(),
		HTTPServer:       initHTTPServerConfig(),
		ServiceDiscovery: initServiceDiscoveryConfig(),
		RateLimit:        initRateLimitConfig(),
		Tracing:          initTracingConfig(),
		Metrics:          initMetricsConfig(),
		CircuitBreaker:   initCircuitBreakerConfig(),
	}

	return cfg, nil
}
