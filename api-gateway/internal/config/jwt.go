package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// JWTConfig holds JWT configuration values.
type JWTConfig struct {
	Secret         string        `mapstructure:"JWT_SECRET"`
	ExpirationTime time.Duration `mapstructure:"JWT_EXPIRATION_TIME"`
	RefreshTime    time.Duration `mapstructure:"JWT_REFRESH_TIME"`
	Issuer         string        `mapstructure:"JWT_ISSUER"`
	TokenLookup    string        `mapstructure:"JWT_TOKEN_LOOKUP"`
	AuthScheme     string        `mapstructure:"JWT_AUTH_SCHEME"`
	SigningMethod  string        `mapstructure:"JWT_SIGNING_METHOD"`
	AllowedAlgs    []string      `mapstructure:"JWT_ALLOWED_ALGS"`
	ContextKey     string        `mapstructure:"JWT_CONTEXT_KEY"`
}

// initJWTConfig initializes the JWT configuration.
func initJWTConfig() *JWTConfig {
	viper.SetDefault("JWT_SECRET", "your-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRATION_TIME", "24h")
	viper.SetDefault("JWT_REFRESH_TIME", "72h")
	viper.SetDefault("JWT_ISSUER", "auth-service")
	viper.SetDefault("JWT_TOKEN_LOOKUP", "header:Authorization")
	viper.SetDefault("JWT_AUTH_SCHEME", "Bearer")
	viper.SetDefault("JWT_SIGNING_METHOD", "HS256")
	viper.SetDefault("JWT_CONTEXT_KEY", "user")
	viper.SetDefault("JWT_ALLOWED_ALGS", []string{"HS256"})

	jwtConfig := &JWTConfig{}

	if err := viper.Unmarshal(&jwtConfig); err != nil {
		log.Fatalf("error mapping jwt config: %v", err)
	}

	return jwtConfig
}
