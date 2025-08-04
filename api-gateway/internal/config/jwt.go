package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret           string        `mapstructure:"JWT_SECRET"`
	ExpirationTime   time.Duration `mapstructure:"JWT_EXPIRATION_TIME"`
	RefreshTime      time.Duration `mapstructure:"JWT_REFRESH_TIME"`
	Issuer           string        `mapstructure:"JWT_ISSUER"`
	TokenLookup      string        `mapstructure:"JWT_TOKEN_LOOKUP"`
	AuthScheme       string        `mapstructure:"JWT_AUTH_SCHEME"`
	SigningMethod    string        `mapstructure:"JWT_SIGNING_METHOD"`
	ContextKey       string        `mapstructure:"JWT_CONTEXT_KEY"`
	TokenLookupQuery string        `mapstructure:"JWT_TOKEN_LOOKUP_QUERY"`
}

// initJWTConfig initializes the JWT configuration from environment variables.
func initJWTConfig() *JWTConfig {
	jwtConfig := &JWTConfig{}

	if err := viper.Unmarshal(&jwtConfig); err != nil {
		log.Fatalf("error mapping jwt config: %v", err)
	}

	return jwtConfig
}
