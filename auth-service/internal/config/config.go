// Package config provides configuration for the auth service.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	App              *AppConfig              `mapstructure:"app"`
	HTTPServer       *HTTPServerConfig       `mapstructure:"http_server"`
	Postgres         *PostgresConfig         `mapstructure:"postgres"`
	JWT              *JWTConfig              `mapstructure:"jwt"`
	ServiceDiscovery *ServiceDiscoveryConfig `mapstructure:"service_discovery"`
	EventPublisher   *EventPublisherConfig   `mapstructure:"event_publisher"`
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := &Config{
		App:              initAppConfig(),
		HTTPServer:       initHTTPServerConfig(),
		Postgres:         initPostgresConfig(),
		JWT:              initJWTConfig(),
		ServiceDiscovery: initServiceDiscoveryConfig(),
		EventPublisher:   initEventPublisherConfig(),
	}

	return config, nil
}
