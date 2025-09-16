package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// AsynqConfig holds the configuration for Asynq task queue.
type AsynqConfig struct {
	// Redis cluster connection settings
	RedisAddrs    []string `mapstructure:"ASYNQ_REDIS_ADDRS"`
	RedisPassword string   `mapstructure:"ASYNQ_REDIS_PASSWORD"`

	// Server settings
	Concurrency int            `mapstructure:"ASYNQ_CONCURRENCY"`
	Queues      map[string]int `mapstructure:"-"` // Set programmatically

	// Retry settings
	MaxRetry      int           `mapstructure:"ASYNQ_MAX_RETRY"`
	RetryDelay    time.Duration `mapstructure:"-"` // Set programmatically
	RetryMaxDelay time.Duration `mapstructure:"-"` // Set programmatically

	// Health check settings
	HealthCheckInterval      time.Duration `mapstructure:"-"` // Set programmatically
	DelayedTaskCheckInterval time.Duration `mapstructure:"-"` // Set programmatically
}

// initAsynqConfig initializes Asynq configuration from environment variables.
func initAsynqConfig() *AsynqConfig {
	// Use same Redis cluster addresses as main Redis config
	viper.SetDefault(
		"ASYNQ_REDIS_ADDRS",
		[]string{
			"localhost:6379", // redis-1 mapped (cluster)
			"localhost:6380", // redis-2 mapped
			"localhost:6381", // redis-3 mapped
			"localhost:6382", // redis-4 mapped
			"localhost:6383", // redis-5 mapped
			"localhost:6384", // redis-6 mapped
		},
	)
	viper.SetDefault("ASYNQ_REDIS_PASSWORD", "supersecret")
	viper.SetDefault("ASYNQ_CONCURRENCY", constant.DefaultAsynqConcurrency)
	viper.SetDefault("ASYNQ_MAX_RETRY", constant.DefaultAsynqMaxRetry)

	config := &AsynqConfig{
		RedisAddrs:    viper.GetStringSlice("ASYNQ_REDIS_ADDRS"),
		RedisPassword: viper.GetString("ASYNQ_REDIS_PASSWORD"),
		Concurrency:   viper.GetInt("ASYNQ_CONCURRENCY"),
		MaxRetry:      viper.GetInt("ASYNQ_MAX_RETRY"),

		// Default queue priorities
		Queues: map[string]int{
			"critical": constant.QueuePriorityCritical, // High priority for urgent tasks
			"default":  constant.QueuePriorityDefault,  // Normal priority
			"low":      constant.QueuePriorityLow,      // Low priority for background tasks
		},

		// Default retry settings
		RetryDelay:    constant.DefaultRetryDelay,
		RetryMaxDelay: constant.DefaultRetryMaxDelay,

		// Default health check settings
		HealthCheckInterval:      constant.DefaultHealthCheckInterval,
		DelayedTaskCheckInterval: constant.DefaultDelayedTaskCheckInterval,
	}

	return config
}
