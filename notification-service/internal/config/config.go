package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App            *AppConfig
	HTTPServer     *HTTPServerConfig
	SSEServer      *SSEServerConfig
	Mail           *MailConfig
	Kafka          *KafkaConfig
	Postgres       *PostgresConfig
	Redis          *RedisConfig
	Consul         *ConsulConfig
	InboxProcessor *InboxProcessorConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	//nolint:errcheck // .env file not required when using environment variables
	_ = viper.ReadInConfig()

	cfg := &Config{
		App:            initAppConfig(),
		Mail:           initMailConfig(),
		HTTPServer:     initHTTPServerConfig(),
		SSEServer:      initSSEServerConfig(),
		Postgres:       initPostgresConfig(),
		Redis:          initRedisConfig(),
		Kafka:          initKafkaConfig(),
		Consul:         initConsulConfig(),
		InboxProcessor: initInboxProcessorConfig(),
	}

	return cfg, nil
}
