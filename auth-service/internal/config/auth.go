package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// AuthConfig holds authentication-specific configuration.
type AuthConfig struct {
	VerificationTokenExpiration time.Duration `mapstructure:"AUTH_VERIFICATION_TOKEN_EXPIRATION"`
}

// initAuthConfig initializes the authentication configuration from environment variables.
func initAuthConfig() *AuthConfig {
	viper.SetDefault("AUTH_VERIFICATION_TOKEN_EXPIRATION", constant.AuthVerificationTokenExpiration)

	authConfig := &AuthConfig{}
	if err := viper.Unmarshal(&authConfig); err != nil {
		panic(err)
	}

	return authConfig
}
