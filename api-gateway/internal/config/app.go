// Package config provides configuration management for the application.
package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultTimeoutShutdown = 10 * time.Second
	defaultTimeoutGateway  = 30 * time.Second
)

// AppConfig holds the application configuration.
type AppConfig struct {
	Name            string        `mapstructure:"APP_NAME"`
	Environment     string        `mapstructure:"APP_ENVIRONMENT"`
	TimeoutShutdown time.Duration `mapstructure:"APP_TIMEOUT_SHUTDOWN"`
	TimeoutGateway  time.Duration `mapstructure:"APP_TIMEOUT_GATEWAY"`
}

// initAppConfig initializes the application configuration from environment variables.
func initAppConfig() *AppConfig {
	// Set defaults
	viper.SetDefault("APP_NAME", "api-gateway")
	viper.SetDefault("APP_ENVIRONMENT", "development")
	viper.SetDefault("APP_TIMEOUT_SHUTDOWN", defaultTimeoutShutdown)
	viper.SetDefault("APP_TIMEOUT_GATEWAY", defaultTimeoutGateway)

	appConfig := &AppConfig{}
	if err := viper.Unmarshal(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
