package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// AsynqConfig holds the configuration for Asynq task queue.
type AsynqConfig struct {
	RedisAddrs               []string       `mapstructure:"ASYNQ_REDIS_ADDRS"`
	RedisPassword            string         `mapstructure:"ASYNQ_REDIS_PASSWORD"`
	Concurrency              int            `mapstructure:"ASYNQ_CONCURRENCY"`
	MaxRetry                 int            `mapstructure:"ASYNQ_MAX_RETRY"`
	RetryDelay               time.Duration  `mapstructure:"ASYNQ_RETRY_DELAY"`
	RetryMaxDelay            time.Duration  `mapstructure:"ASYNQ_RETRY_MAX_DELAY"`
	HealthCheckInterval      time.Duration  `mapstructure:"ASYNQ_HEALTH_CHECK_INTERVAL"`
	DelayedTaskCheckInterval time.Duration  `mapstructure:"ASYNQ_DELAYED_TASK_CHECK_INTERVAL"`
	Queues                   map[string]int `mapstructure:"-"`
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
	viper.SetDefault("ASYNQ_RETRY_DELAY", constant.DefaultRetryDelay)
	viper.SetDefault("ASYNQ_RETRY_MAX_DELAY", constant.DefaultRetryMaxDelay)
	viper.SetDefault("ASYNQ_HEALTH_CHECK_INTERVAL", constant.DefaultHealthCheckInterval)
	viper.SetDefault("ASYNQ_DELAYED_TASK_CHECK_INTERVAL", constant.DefaultDelayedTaskCheckInterval)
	viper.SetDefault("ASYNQ_QUEUE_CRITICAL_PRIORITY", constant.QueuePriorityCritical)
	viper.SetDefault("ASYNQ_QUEUE_DEFAULT_PRIORITY", constant.QueuePriorityDefault)
	viper.SetDefault("ASYNQ_QUEUE_LOW_PRIORITY", constant.QueuePriorityLow)

	asynqConfig := &AsynqConfig{}
	if err := viper.Unmarshal(asynqConfig); err != nil {
		panic(err)
	}

	// Parse comma-separated ASYNQ_REDIS_ADDRS string from environment variable
	addrsStr := viper.GetString("ASYNQ_REDIS_ADDRS")
	if addrsStr != "" {
		asynqConfig.RedisAddrs = parseCommaSeparatedAsynq(addrsStr)
	}

	asynqConfig.Queues = map[string]int{
		"critical": viper.GetInt("ASYNQ_QUEUE_CRITICAL_PRIORITY"),
		"default":  viper.GetInt("ASYNQ_QUEUE_DEFAULT_PRIORITY"),
		"low":      viper.GetInt("ASYNQ_QUEUE_LOW_PRIORITY"),
	}

	return asynqConfig
}

// parseCommaSeparatedAsynq parses a comma-separated string into a slice of strings.
func parseCommaSeparatedAsynq(s string) []string {
	parts := strings.Split(s, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
