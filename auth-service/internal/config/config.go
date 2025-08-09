// Package config provides configuration for the auth service.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	App        *AppConfig        `mapstructure:"app"`
	HTTPServer *HTTPServerConfig `mapstructure:"http_server"`
	JWT        *JWTConfig        `mapstructure:"jwt"`
	Postgres   *PostgresConfig   `mapstructure:"postgres"`
	Kafka      *KafkaConfig      `mapstructure:"kafka"`
	Consul     *ConsulConfig     `mapstructure:"consul"`
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := &Config{
		App:        initAppConfig(),
		HTTPServer: initHTTPServerConfig(),
		Postgres:   initPostgresConfig(),
		Kafka:      initKafkaConfig(),
		JWT:        initJWTConfig(),
		Consul:     initConsulConfig(),
	}

	return config, nil
}
