package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Addrs           []string      `mapstructure:"REDIS_ADDRS"`
	Password        string        `mapstructure:"REDIS_PASSWORD"`
	DialTimeout     time.Duration `mapstructure:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout     time.Duration `mapstructure:"REDIS_READ_TIMEOUT"`
	WriteTimeout    time.Duration `mapstructure:"REDIS_WRITE_TIMEOUT"`
	MinIdleConn     int           `mapstructure:"REDIS_MIN_IDLE_CONN"`
	MaxIdleConn     int           `mapstructure:"REDIS_MAX_IDLE_CONN"`
	MaxActiveConn   int           `mapstructure:"REDIS_MAX_ACTIVE_CONN"`
	MaxConnLifetime time.Duration `mapstructure:"REDIS_MAX_CONN_LIFETIME"`
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
	viper.SetDefault("REDIS_DIAL_TIMEOUT", constant.RedisDialTimeout)
	viper.SetDefault("REDIS_READ_TIMEOUT", constant.RedisReadTimeout)
	viper.SetDefault("REDIS_WRITE_TIMEOUT", constant.RedisWriteTimeout)
	viper.SetDefault("REDIS_MIN_IDLE_CONN", constant.RedisMinIdleConn)
	viper.SetDefault("REDIS_MAX_IDLE_CONN", constant.RedisMaxIdleConn)
	viper.SetDefault("REDIS_MAX_ACTIVE_CONN", constant.RedisMaxActiveConn)
	viper.SetDefault("REDIS_MAX_CONN_LIFETIME", constant.RedisMaxConnLifetime)

	redisConfig := &RedisConfig{}
	if err := viper.Unmarshal(redisConfig); err != nil {
		panic(err)
	}

	// Parse comma-separated REDIS_ADDRS string from environment variable
	addrsStr := viper.GetString("REDIS_ADDRS")
	if addrsStr != "" {
		redisConfig.Addrs = parseCommaSeparated(addrsStr)
	}

	return redisConfig
}
