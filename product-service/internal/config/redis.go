package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultRedisDialTimeout     = 5 * time.Second
	defaultRedisReadTimeout     = 3 * time.Second
	defaultRedisWriteTimeout    = 3 * time.Second
	defaultRedisMinIdleConn     = 10
	defaultRedisMaxIdleConn     = 100
	defaultRedisMaxActiveConn   = 100
	defaultRedisMaxConnLifetime = 5 * time.Minute
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
	viper.SetDefault("REDIS_DIAL_TIMEOUT", defaultRedisDialTimeout)
	viper.SetDefault("REDIS_READ_TIMEOUT", defaultRedisReadTimeout)
	viper.SetDefault("REDIS_WRITE_TIMEOUT", defaultRedisWriteTimeout)
	viper.SetDefault("REDIS_MIN_IDLE_CONN", defaultRedisMinIdleConn)
	viper.SetDefault("REDIS_MAX_IDLE_CONN", defaultRedisMaxIdleConn)
	viper.SetDefault("REDIS_MAX_ACTIVE_CONN", defaultRedisMaxActiveConn)
	viper.SetDefault("REDIS_MAX_CONN_LIFETIME", defaultRedisMaxConnLifetime)

	redisConfig := &RedisConfig{}
	if err := viper.Unmarshal(redisConfig); err != nil {
		panic(err)
	}

	return redisConfig
}
