package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// OutboxPublisherConfig holds OutboxPublisher service discovery configuration.
type OutboxPublisherConfig struct {
	BatchSize        int           `mapstructure:"OUTBOX_BATCH_SIZE"`
	PollInterval     time.Duration `mapstructure:"OUTBOX_POLL_INTERVAL"`
	MaxRetryAttempts int           `mapstructure:"OUTBOX_MAX_RETRY_ATTEMPTS"`
	RetryBackoff     time.Duration `mapstructure:"OUTBOX_RETRY_BACKOFF"`
	CleanupInterval  time.Duration `mapstructure:"OUTBOX_CLEANUP_INTERVAL"`
	RetentionPeriod  time.Duration `mapstructure:"OUTBOX_RETENTION_PERIOD"`
}

// initOutboxPublisherConfig initializes the OutboxPublisher configuration from environment variables.
func initOutboxPublisherConfig() *OutboxPublisherConfig {
	// Set defaults
	viper.SetDefault("OUTBOX_BATCH_SIZE", 100)
	viper.SetDefault("OUTBOX_POLL_INTERVAL", 5*time.Second)
	viper.SetDefault("OUTBOX_MAX_RETRY_ATTEMPTS", 5)
	viper.SetDefault("OUTBOX_RETRY_BACKOFF", 30*time.Second)
	viper.SetDefault("OUTBOX_CLEANUP_INTERVAL", 1*time.Hour)
	viper.SetDefault("OUTBOX_RETENTION_PERIOD", 24*time.Hour)

	outboxConfig := &OutboxPublisherConfig{}

	if err := viper.Unmarshal(&outboxConfig); err != nil {
		log.Fatalf("error mapping outbox config: %v", err)
	}

	return outboxConfig
}
