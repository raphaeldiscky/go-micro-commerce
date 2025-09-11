package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// AppConfig holds application configuration.
type AppConfig struct {
	Name        string `mapstructure:"APP_NAME"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	// Set defaults
	viper.SetDefault("APP_NAME", "auth-service")
	viper.SetDefault("APP_ENVIRONMENT", "development")

	appConfig := &AppConfig{}

	if err := viper.Unmarshal(&appConfig); err != nil {
		slog.Error("error mapping app config", "err", err)
	}

	return appConfig
}
