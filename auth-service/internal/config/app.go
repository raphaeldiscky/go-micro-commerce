package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig holds application configuration.
type AppConfig struct {
	Name        string `mapstructure:"APP_NAME"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`
	Version     string `mapstructure:"APP_VERSION"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	// Set defaults
	viper.SetDefault("APP_NAME", "auth-service")
	viper.SetDefault("APP_ENVIRONMENT", "development")
	viper.SetDefault("APP_VERSION", "1.0.0")
	viper.SetDefault("LOG_LEVEL", "info")

	appConfig := &AppConfig{}

	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("error mapping app config: %v", err)
	}

	return appConfig
}
