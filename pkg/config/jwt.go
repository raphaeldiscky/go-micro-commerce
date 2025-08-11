package config

import (
	"log"

	"github.com/spf13/viper"
)

// JWTConfig holds JWT configuration values.
type JWTConfig struct {
	AllowedAlgs   []string `mapstructure:"JWT_ALLOWED_ALGS"`
	Issuer        string   `mapstructure:"JWT_ISSUER"`
	SecretKey     string   `mapstructure:"JWT_SECRET_KEY"`
	TokenDuration int      `mapstructure:"JWT_TOKEN_DURATION"`
}

// initJWTConfig initializes the JWT configuration.
func initJWTConfig() *JWTConfig {
	// Set defaults
	viper.SetDefault("JWT_ALLOWED_ALGS", []string{"HS256"})
	viper.SetDefault("JWT_ISSUER", "example.com")
	viper.SetDefault("JWT_SECRET_KEY", "supersecretkey")
	viper.SetDefault("JWT_TOKEN_DURATION", 3600)

	jwtConfig := &JWTConfig{}

	if err := viper.Unmarshal(&jwtConfig); err != nil {
		log.Fatalf("error mapping jwt config: %v", err)
	}

	return jwtConfig
}
