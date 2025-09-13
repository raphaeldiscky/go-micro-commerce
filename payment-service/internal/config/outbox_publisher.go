package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// OutboxPublisherConfig holds OutboxPublisher service discovery configuration.
type OutboxPublisherConfig struct {
	BatchSize        int64         `mapstructure:"OUTBOX_BATCH_SIZE"`
	PollInterval     time.Duration `mapstructure:"OUTBOX_POLL_INTERVAL"`
	MaxRetryAttempts int64         `mapstructure:"OUTBOX_MAX_RETRY_ATTEMPTS"`
	RetryBackoff     time.Duration `mapstructure:"OUTBOX_RETRY_BACKOFF"`
	CleanupInterval  time.Duration `mapstructure:"OUTBOX_CLEANUP_INTERVAL"`
	RetentionPeriod  time.Duration `mapstructure:"OUTBOX_RETENTION_PERIOD"`
}

// initOutboxPublisherConfig initializes the OutboxPublisher configuration from environment variables.
func initOutboxPublisherConfig() *OutboxPublisherConfig {
	// Set defaults
	viper.SetDefault("OUTBOX_BATCH_SIZE", constant.OutboxBatchSize)
	viper.SetDefault("OUTBOX_POLL_INTERVAL", constant.OutboxPollInterval)
	viper.SetDefault("OUTBOX_MAX_RETRY_ATTEMPTS", constant.OutboxMaxRetryAttempts)
	viper.SetDefault("OUTBOX_RETRY_BACKOFF", constant.OutboxRetryBackoff)
	viper.SetDefault("OUTBOX_CLEANUP_INTERVAL", constant.OutboxCleanupInterval)
	viper.SetDefault("OUTBOX_RETENTION_PERIOD", constant.OutboxRetentionPeriod)

	outboxConfig := &OutboxPublisherConfig{}

	if err := viper.Unmarshal(outboxConfig); err != nil {
		panic(err)
	}

	return outboxConfig
}
