// Package config provides configuration management for the API gateway.
package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App              *AppConfig
	HTTPServer       *HTTPServerConfig
	JWT              *JWTConfig
	ServiceDiscovery *ServiceDiscoveryConfig
	RateLimit        *RateLimitConfig
	Tracing          *TracingConfig
	Metrics          *MetricsConfig
}

// LoadConfig loads configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	configPath := parseConfigPath()
	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	cfg := &Config{
		App:              initAppConfig(),
		HTTPServer:       initHTTPServerConfig(),
		JWT:              initJWTConfig(),
		ServiceDiscovery: initServiceDiscoveryConfig(),
		RateLimit:        initRateLimitConfig(),
		Tracing:          initTracingConfig(),
		Metrics:          initMetricsConfig(),
	}

	return cfg, nil
}

// parseConfigPath returns the current working directory.
func parseConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd
}
