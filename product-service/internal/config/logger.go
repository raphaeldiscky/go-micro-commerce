package config

import (
	"log"

	"github.com/spf13/viper"
)

// LoggerConfig holds the configuration for the logger.
type LoggerConfig struct {
	Level int `mapstructure:"LOGGER_LEVEL"`
}

// LoggerConfig holds the configuration for the logger.
func initLoggerConfig() *LoggerConfig {
	loggerConfig := &LoggerConfig{}

	if err := viper.Unmarshal(&loggerConfig); err != nil {
		log.Fatalf("error mapping logger config: %v", err)
	}

	return loggerConfig
}
