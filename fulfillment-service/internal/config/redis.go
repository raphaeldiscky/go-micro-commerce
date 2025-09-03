package config

import (
	"log"

	"github.com/spf13/viper"
)

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addrs           []string `mapstructure:"REDIS_ADDRS"`
	Password        string   `mapstructure:"REDIS_PASSWORD"`
	DialTimeout     int      `mapstructure:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout     int      `mapstructure:"REDIS_READ_TIMEOUT"`
	WriteTimeout    int      `mapstructure:"REDIS_WRITE_TIMEOUT"`
	MinIdleConn     int      `mapstructure:"REDIS_MIN_IDLE_CONN"`
	MaxIdleConn     int      `mapstructure:"REDIS_MAX_IDLE_CONN"`
	MaxActiveConn   int      `mapstructure:"REDIS_MAX_ACTIVE_CONN"`
	MaxConnLifetime int      `mapstructure:"REDIS_MAX_CONN_LIFETIME"`
}

// initRedisConfig initializes the Redis configuration from environment variables.
func initRedisConfig() *RedisConfig {
	// Set defaults
	viper.SetDefault(
		"REDIS_ADDRS",
		[]string{
			"localhost:6379", // redis-1 mapped (cluster)
			"localhost:6380", // redis-2 mapped
			"localhost:6381", // redis-3 mapped
			"localhost:6382", // redis-4 mapped
			"localhost:6383", // redis-5 mapped
			"localhost:6384", // redis-6 mapped
		},
	)
	viper.SetDefault("REDIS_PASSWORD", "supersecret")
	viper.SetDefault("REDIS_DIAL_TIMEOUT", 5)
	viper.SetDefault("REDIS_READ_TIMEOUT", 3)
	viper.SetDefault("REDIS_WRITE_TIMEOUT", 3)
	viper.SetDefault("REDIS_MIN_IDLE_CONN", 10)
	viper.SetDefault("REDIS_MAX_IDLE_CONN", 100)
	viper.SetDefault("REDIS_MAX_ACTIVE_CONN", 100)
	viper.SetDefault("REDIS_MAX_CONN_LIFETIME", 60)

	redisConfig := &RedisConfig{}

	if err := viper.Unmarshal(&redisConfig); err != nil {
		log.Fatalf("error mapping Redis config: %v", err)
	}

	return redisConfig
}
