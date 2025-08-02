// Package config provides configuration management for the application.
package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	// Paseto holds the PASETO token configuration.
	Paseto *PasetoConfig
	// SMTP holds the SMTP server configuration.
	SMTP *SMTPConfig
}

// NewConfig creates a new configuration instance by loading environment variables
// and setting up PASETO and SMTP configurations.
func NewConfig() (*Config, error) {
	configPath := parseConfigPath()
	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Config{
		Paseto: initPasetoConfig(),
		SMTP:   initSMTPConfig(),
	}, nil
}

// InitConfig initializes the configuration by loading environment variables
// and setting up PASETO and SMTP configurations.
// Deprecated: Use NewConfig instead for better error handling.
func InitConfig() (*Config, error) {
	return NewConfig()
}

func parseConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return wd
}
