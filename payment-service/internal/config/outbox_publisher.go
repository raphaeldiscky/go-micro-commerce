package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultOutboxBatchSize        = 100
	defaultOutboxPollInterval     = 5 * time.Second
	defaultOutboxMaxRetryAttempts = 5
	defaultOutboxRetryBackoff     = 30 * time.Second
	defaultOutboxCleanupInterval  = 1 * time.Hour
	defaultOutboxRetentionPeriod  = 24 * time.Hour
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
	viper.SetDefault("OUTBOX_BATCH_SIZE", defaultOutboxBatchSize)
	viper.SetDefault("OUTBOX_POLL_INTERVAL", defaultOutboxPollInterval)
	viper.SetDefault("OUTBOX_MAX_RETRY_ATTEMPTS", defaultOutboxMaxRetryAttempts)
	viper.SetDefault("OUTBOX_RETRY_BACKOFF", defaultOutboxRetryBackoff)
	viper.SetDefault("OUTBOX_CLEANUP_INTERVAL", defaultOutboxCleanupInterval)
	viper.SetDefault("OUTBOX_RETENTION_PERIOD", defaultOutboxRetentionPeriod)

	outboxConfig := &OutboxPublisherConfig{}

	if err := viper.Unmarshal(outboxConfig); err != nil {
		panic(err)
	}

	return outboxConfig
}
