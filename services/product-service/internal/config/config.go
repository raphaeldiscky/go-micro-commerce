package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App        *AppConfig
	HTTPServer *HTTPServerConfig
	Postgres   *PostgresConfig
	Kafka      *KafkaConfig
	Logger     *LoggerConfig
	Redis      *RedisConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	configPath := parseConfigPath()
	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	cfg := &Config{
		App:        initAppConfig(),
		HTTPServer: initHTTPServerConfig(),
		Postgres:   initPostgresConfig(),
		Kafka:      initKafkaConfig(),
		Logger:     initLoggerConfig(),
		Redis:      initRedisConfig(),
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
