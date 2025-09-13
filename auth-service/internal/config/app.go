// Package config provides configuration management for the application.
package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Name            string        `mapstructure:"APP_NAME"`
	Environment     string        `mapstructure:"APP_ENVIRONMENT"`
	LoggerLevel     int           `mapstructure:"APP_LOGGER_LEVEL"`
	TimeoutShutdown time.Duration `mapstructure:"APP_SHUTDOWN_TIMEOUT"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	// Set defaults
	viper.SetDefault("APP_NAME", "auth-service")
	viper.SetDefault("APP_ENVIRONMENT", "development")
	viper.SetDefault("APP_LOGGER_LEVEL", constant.AppLoggerLevel)
	viper.SetDefault("APP_TIMEOUT_SHUTDOWN", constant.AppTimeoutShutdown)

	appConfig := &AppConfig{}
	if err := viper.Unmarshal(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
