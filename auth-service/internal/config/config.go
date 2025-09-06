// Package config provides configuration for the auth service.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	App        *AppConfig
	Logger     *LoggerConfig
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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := &Config{
		App:        initAppConfig(),
		Logger:     initLoggerConfig(),
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
