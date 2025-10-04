package config

import (
	"time"

	"github.com/spf13/viper"
)

// JWTConfig holds JWT configuration values.
type JWTConfig struct {
	Secret         string        `mapstructure:"JWT_SECRET"`
	PublicKeyPath  string        `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	PrivateKeyPath string        `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	ExpirationTime time.Duration `mapstructure:"JWT_EXPIRATION_TIME"`
	RefreshTime    time.Duration `mapstructure:"JWT_REFRESH_TIME"`
	Issuer         string        `mapstructure:"JWT_ISSUER"`
	TokenLookup    string        `mapstructure:"JWT_TOKEN_LOOKUP"`
	AuthScheme     string        `mapstructure:"JWT_AUTH_SCHEME"`
	SigningMethod  string        `mapstructure:"JWT_SIGNING_METHOD"`
	ContextKey     string        `mapstructure:"JWT_CONTEXT_KEY"`
	AllowedAlgs    []string      `mapstructure:"JWT_ALLOWED_ALGS"`
}

// initJWTConfig initializes the JWT configuration.
func initJWTConfig() *JWTConfig {
	// Set defaults
	viper.SetDefault("JWT_SECRET", "secret")
	viper.SetDefault("JWT_EXPIRATION_TIME", "24h")
	viper.SetDefault("JWT_REFRESH_TIME", "72h")
	viper.SetDefault("JWT_ISSUER", "auth-service")
	viper.SetDefault("JWT_TOKEN_LOOKUP", "header:Authorization")
	viper.SetDefault("JWT_AUTH_SCHEME", "Bearer")
	viper.SetDefault("JWT_SIGNING_METHOD", "HS256")
	viper.SetDefault("JWT_CONTEXT_KEY", "user")
	viper.SetDefault("JWT_ALLOWED_ALGS", []string{"HS256"})

	jwtConfig := &JWTConfig{}
	if err := viper.Unmarshal(jwtConfig); err != nil {
		panic(err)
	}

	return jwtConfig
}
