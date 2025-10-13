package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App            *AppConfig
	HTTPServer     *HTTPServerConfig
	SSEServer      *SSEServerConfig
	SMTP           *SMTPConfig
	Kafka          *KafkaConfig
	Postgres       *PostgresConfig
	Redis          *RedisConfig
	Consul         *ConsulConfig
	InboxProcessor *InboxProcessorConfig
	Sharding       *ShardingConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	cfg := &Config{
		App:            initAppConfig(),
		SMTP:           initSMTPConfig(),
		HTTPServer:     initHTTPServerConfig(),
		SSEServer:      initSSEServerConfig(),
		Postgres:       initPostgresConfig(),
		Redis:          initRedisConfig(),
		Kafka:          initKafkaConfig(),
		Consul:         initConsulConfig(),
		InboxProcessor: initInboxProcessorConfig(),
		Sharding:       initShardingConfig(),
	}

	return cfg, nil
}
