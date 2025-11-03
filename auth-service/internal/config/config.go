// Package config provides configuration for the auth service.
package config

import (
	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	App        *AppConfig
	HTTPServer *HTTPServerConfig
	JWT        *JWTConfig
	Bcrypt     *BcryptConfig
	Auth       *AuthConfig
	Postgres   *PostgresConfig
	Kafka      *KafkaConfig
	Consul     *ConsulConfig
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	//nolint:errcheck // .env file not required when using environment variables
	_ = viper.ReadInConfig()

	config := &Config{
		App:        initAppConfig(),
		HTTPServer: initHTTPServerConfig(),
		Postgres:   initPostgresConfig(),
		Kafka:      initKafkaConfig(),
		JWT:        initJWTConfig(),
		Bcrypt:     initBcryptConfig(),
		Auth:       initAuthConfig(),
		Consul:     initConsulConfig(),
	}

	return config, nil
}
