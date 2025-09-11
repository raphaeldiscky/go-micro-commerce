package config

import (
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

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
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
