// Package config provides configuration management for the application.
package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	JWT        *JWTConfig
	SMTP       *SMTPConfig
	GRPCClient *GRPCClientConfig
	Kafka      *KafkaConfig
	Asynq      *AsynqConfig
}

// NewConfig creates a new configuration instance by loading environment variables.
func NewConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Config{
		JWT:        initJWTConfig(),
		SMTP:       initSMTPConfig(),
		GRPCClient: initGRPCClientConfig("api"),
		Kafka:      initKafkaConfig(),
		Asynq:      initAsynqConfig(),
	}, nil
}
