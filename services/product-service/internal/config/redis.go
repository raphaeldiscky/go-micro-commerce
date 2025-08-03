package config

import (
	"log"

	"github.com/spf13/viper"
)

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// initRedisConfig initializes the Redis configuration from environment variables.
func initRedisConfig() *RedisConfig {
	redisConfig := &RedisConfig{}

	if err := viper.Unmarshal(&redisConfig); err != nil {
		log.Fatalf("error mapping Redis config: %v", err)
	}

	return redisConfig
}
