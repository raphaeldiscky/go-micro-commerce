package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App             *AppConfig
	HTTPServer      *HTTPServerConfig
	GRPCServer      *GRPCServerConfig
	Postgres        *PostgresConfig
	Kafka           *KafkaConfig
	Redis           *RedisConfig
	Consul          *ConsulConfig
	OutboxPublisher *OutboxPublisherConfig
	Client          *ClientConfig
	Asynq           *AsynqConfig
	Tracing         *TracingConfig
	Metrics         *MetricsConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	//nolint:errcheck // .env file not required when using environment variables
	_ = viper.ReadInConfig()

	cfg := &Config{
		App:             initAppConfig(),
		HTTPServer:      initHTTPServerConfig(),
		GRPCServer:      initGRPCServerConfig(),
		Postgres:        initPostgresConfig(),
		Kafka:           initKafkaConfig(),
		Redis:           initRedisConfig(),
		Consul:          initConsulConfig(),
		OutboxPublisher: initOutboxPublisherConfig(),
		Client:          initClientConfig(),
		Asynq:           initAsynqConfig(),
		Tracing:         initTracingConfig(),
		Metrics:         initMetricsConfig(),
	}

	return cfg, nil
}
