package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App             *AppConfig
	Logger          *LoggerConfig
	HTTPServer      *HTTPServerConfig
	Postgres        *PostgresConfig
	Kafka           *KafkaConfig
	Redis           *RedisConfig
	Consul          *ConsulConfig
	OutboxPublisher *OutboxPublisherConfig
	Client          *ClientConfig
	Saga            *SagaConfig
	Jobs            *JobsConfig
	Temporal        *TemporalConfig
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
		App:             initAppConfig(),
		Logger:          initLoggerConfig(),
		HTTPServer:      initHTTPServerConfig(),
		Postgres:        initPostgresConfig(),
		Kafka:           initKafkaConfig(),
		Redis:           initRedisConfig(),
		Consul:          initConsulConfig(),
		OutboxPublisher: initOutboxPublisherConfig(),
		Client:          initClientConfig(),
		Saga:            initSagaConfig(),
		Jobs:            initJobsConfig(),
		Temporal:        initTemporalConfig(),
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
