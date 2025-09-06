package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App            *AppConfig
	Logger         *LoggerConfig
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

	configPath := parseConfigPath()
	viper.SetConfigFile(configPath + "/.env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	cfg := &Config{
		App:            initAppConfig(),
		Logger:         initLoggerConfig(),
		SMTP:           initSMTPConfig(),
		HTTPServer:     initHTTPServerConfig(),
		Postgres:       initPostgresConfig(),
		Kafka:          initKafkaConfig(),
		Consul:         initConsulConfig(),
		InboxProcessor: initInboxProcessorConfig(),
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
