package config

import (
	"log"

	"github.com/spf13/viper"
)

// BcryptConfig holds Bcrypt configuration.
type BcryptConfig struct {
	Cost int `mapstructure:"BCRYPT_COST"`
}

// initBcryptConfig initializes the Bcrypt configuration from environment variables.
func initBcryptConfig() *BcryptConfig {
	// Set defaults
	viper.SetDefault("BCRYPT_COST", 10)

	bcryptConfig := &BcryptConfig{}

	if err := viper.Unmarshal(&bcryptConfig); err != nil {
		log.Fatalf("error mapping bcrypt config: %v", err)
	}

	return bcryptConfig
}
