// Package config provides configuration management for the application.
package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Name        string `mapstructure:"APP_NAME"`
	Environment string `mapstructure:"APP_ENVIRONMENT"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	// Set defaults
	viper.SetDefault("APP_NAME", "product-service")
	viper.SetDefault("APP_ENVIRONMENT", "development")

	appConfig := &AppConfig{}

	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("error mapping app config: %v", err)
	}

	return appConfig
}
