package config

import (
	"github.com/spf13/viper"
)

const (
	defaultLoggerLevel = 1
)

// LoggerConfig holds the logger configuration.
type LoggerConfig struct {
	Level int `mapstructure:"LOGGER_LEVEL"`
}

// initLoggerConfig initializes the logger configuration.
func initLoggerConfig() *LoggerConfig {
	// Set defaults
	viper.SetDefault("LOGGER_LEVEL", defaultLoggerLevel)

	loggerConfig := &LoggerConfig{}

	return loggerConfig
}
