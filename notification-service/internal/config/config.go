package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App            *AppConfig
	HTTPServer     *HTTPServerConfig
	SMTP           *SMTPConfig
	Kafka          *KafkaConfig
	Postgres       *PostgresConfig
	Consul         *ConsulConfig
	InboxProcessor *InboxProcessorConfig
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
		Postgres:       initPostgresConfig(),
		Kafka:          initKafkaConfig(),
		Consul:         initConsulConfig(),
		InboxProcessor: initInboxProcessorConfig(),
	}

	return cfg, nil
}
