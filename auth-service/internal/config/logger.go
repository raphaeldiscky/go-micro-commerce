package config

import (
	"log"

	"github.com/spf13/viper"
)

// LoggerConfig holds the logger configuration.
type LoggerConfig struct {
	Level int `mapstructure:"LOGGER_LEVEL"`
}

// initLoggerConfig initializes the logger configuration.
func initLoggerConfig() *LoggerConfig {
	// Set defaults
	viper.SetDefault("LOGGER_LEVEL", 1)

	loggerConfig := &LoggerConfig{}

	if err := viper.Unmarshal(&loggerConfig); err != nil {
		log.Fatalf("error mapping logger config: %v", err)
	}

	return loggerConfig
}
