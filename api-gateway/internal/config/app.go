// Package config provides configuration management for the application.
package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Name        string `mapstructure:"APP_NAME"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	viper.SetDefault("APP_NAME", "api-gateway")
	viper.SetDefault("APP_ENVIRONMENT", "development")

	appConfig := &AppConfig{}

	if err := viper.Unmarshal(&appConfig); err != nil {
		slog.Error("error mapping app config", "err", err)
	}

	return appConfig
}
