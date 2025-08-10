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
	jwtConfig := &JWTConfig{}

	if err := viper.Unmarshal(&jwtConfig); err != nil {
		log.Fatalf("error mapping jwt config: %v", err)
	}

	return jwtConfig
}
