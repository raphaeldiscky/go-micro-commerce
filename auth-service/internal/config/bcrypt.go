package config

import (
	"github.com/spf13/viper"
)

const (
	defaultBcryptCost = 10
)

// BcryptConfig holds Bcrypt configuration.
type BcryptConfig struct {
	Cost int `mapstructure:"BCRYPT_COST"`
}

// initBcryptConfig initializes the Bcrypt configuration from environment variables.
func initBcryptConfig() *BcryptConfig {
	// Set defaults
	viper.SetDefault("BCRYPT_COST", defaultBcryptCost)

	bcryptConfig := &BcryptConfig{}

	if err := viper.Unmarshal(&bcryptConfig); err != nil {
		panic(err)
	}

	return bcryptConfig
}
