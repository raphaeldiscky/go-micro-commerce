package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// AuthConfig holds authentication-specific configuration.
type AuthConfig struct {
	// VerificationTokenExpiration is the duration for which email verification tokens are valid
	VerificationTokenExpiration time.Duration `mapstructure:"AUTH_VERIFICATION_TOKEN_EXPIRATION"`
}

// initAuthConfig initializes the authentication configuration from environment variables.
func initAuthConfig() *AuthConfig {
	// Set default: 10 minutes to match what's stated in the email template
	viper.SetDefault("AUTH_VERIFICATION_TOKEN_EXPIRATION", "10m")

	authConfig := &AuthConfig{}

	if err := viper.Unmarshal(&authConfig); err != nil {
		log.Fatalf("error mapping auth config: %v", err)
	}

	return authConfig
}
