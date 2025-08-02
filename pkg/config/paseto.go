package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// PasetoConfig holds the PASETO token configuration.
type PasetoConfig struct {
	SecretKey       string        `mapstructure:"PASETO_SECRET_KEY"`
	AccessDuration  time.Duration `mapstructure:"PASETO_ACCESS_DURATION"`
	RefreshDuration time.Duration `mapstructure:"PASETO_REFRESH_DURATION"`
	Issuer          string        `mapstructure:"PASETO_ISSUER"`
}

// initPasetoConfig returns the PASETO configuration.
func initPasetoConfig() *PasetoConfig {
	pasetoConfig := &PasetoConfig{}

	if err := viper.Unmarshal(&pasetoConfig); err != nil {
		log.Fatalf("error mapping paseto config: %v", err)
	}

	// Set default values if not provided
	if pasetoConfig.AccessDuration == 0 {
		pasetoConfig.AccessDuration = 15 * time.Minute
	}

	if pasetoConfig.RefreshDuration == 0 {
		pasetoConfig.RefreshDuration = 24 * time.Hour
	}

	if pasetoConfig.Issuer == "" {
		pasetoConfig.Issuer = "marketplace-api"
	}

	return pasetoConfig
}
