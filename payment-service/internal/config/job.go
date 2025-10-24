package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// JobConfig holds configuration for background jobs.
type JobConfig struct {
	PaymentTimeout *PaymentTimeoutJobConfig
}

// PaymentTimeoutJobConfig holds configuration for payment timeout job.
type PaymentTimeoutJobConfig struct {
	Enabled             bool
	Interval            time.Duration
	Timeout             time.Duration
	BatchSize           int
	RedisLockTTL        time.Duration
	RedisLockBackoff    time.Duration
	RedisLockMaxRetries int
}

// initJobConfig initializes job configuration from environment variables.
func initJobConfig() *JobConfig {
	viper.SetDefault("JOB_PAYMENT_TIMEOUT_ENABLED", true)
	viper.SetDefault(
		"JOB_PAYMENT_TIMEOUT_INTERVAL",
		"5m",
	) // Run every 5 minutes
	viper.SetDefault(
		"JOB_PAYMENT_TIMEOUT_TIMEOUT",
		"2m",
	) // Max 2 minutes per execution
	viper.SetDefault(
		"JOB_PAYMENT_TIMEOUT_BATCH_SIZE",
		constant.JobPaymentTimeoutBatchSize,
	) // Process 100 payments per batch
	viper.SetDefault("JOB_PAYMENT_TIMEOUT_REDIS_LOCK_TTL", "3m")
	viper.SetDefault("JOB_PAYMENT_TIMEOUT_REDIS_LOCK_BACKOFF", "500ms")
	viper.SetDefault(
		"JOB_PAYMENT_TIMEOUT_REDIS_LOCK_MAX_RETRIES",
		constant.JobPaymentTimeoutRedisLockMaxRetries,
	)

	return &JobConfig{
		PaymentTimeout: &PaymentTimeoutJobConfig{
			Enabled:             viper.GetBool("JOB_PAYMENT_TIMEOUT_ENABLED"),
			Interval:            viper.GetDuration("JOB_PAYMENT_TIMEOUT_INTERVAL"),
			Timeout:             viper.GetDuration("JOB_PAYMENT_TIMEOUT_TIMEOUT"),
			BatchSize:           viper.GetInt("JOB_PAYMENT_TIMEOUT_BATCH_SIZE"),
			RedisLockTTL:        viper.GetDuration("JOB_PAYMENT_TIMEOUT_REDIS_LOCK_TTL"),
			RedisLockBackoff:    viper.GetDuration("JOB_PAYMENT_TIMEOUT_REDIS_LOCK_BACKOFF"),
			RedisLockMaxRetries: viper.GetInt("JOB_PAYMENT_TIMEOUT_REDIS_LOCK_MAX_RETRIES"),
		},
	}
}
