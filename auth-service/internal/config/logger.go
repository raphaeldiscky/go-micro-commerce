package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// LoggerConfig holds the logger configuration.
type LoggerConfig struct {
	Level int `mapstructure:"LOGGER_LEVEL"`
}

// initLoggerConfig initializes the logger configuration.
func initLoggerConfig() *LoggerConfig {
	// Set defaults
	viper.SetDefault("LOGGER_LEVEL", 4)

	loggerConfig := &LoggerConfig{}

	if err := viper.Unmarshal(&loggerConfig); err != nil {
		slog.Error("error mapping logger config", "err", err)
	}

	return loggerConfig
}
