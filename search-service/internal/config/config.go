package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App            *AppConfig
	Logger         *LoggerConfig
	HTTPServer     *HTTPServerConfig
	Kafka          *KafkaConfig
	Postgres       *PostgresConfig
	Consul         *ConsulConfig
	InboxProcessor *InboxProcessorConfig
	Elasticsearch  *ElasticsearchConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	configPath := parseConfigPath()
	viper.SetConfigFile(configPath + "/.env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		App:            initAppConfig(),
		Logger:         initLoggerConfig(),
		HTTPServer:     initHTTPServerConfig(),
		Postgres:       initPostgresConfig(),
		Kafka:          initKafkaConfig(),
		Consul:         initConsulConfig(),
		InboxProcessor: initInboxProcessorConfig(),
		Elasticsearch:  initElasticsearchConfig(),
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
