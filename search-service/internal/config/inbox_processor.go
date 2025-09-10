package config

import (
	"time"

	"github.com/spf13/viper"
)

// InboxProcessorConfig holds the configuration for the inbox processor.
type InboxProcessorConfig struct {
	PollInterval     time.Duration `mapstructure:"INBOX_POLL_INTERVAL"`
	CleanupInterval  time.Duration `mapstructure:"INBOX_CLEANUP_INTERVAL"`
	RetentionPeriod  time.Duration `mapstructure:"INBOX_RETENTION_PERIOD"`
	BatchSize        int64         `mapstructure:"INBOX_BATCH_SIZE"`
	MaxRetryAttempts int64         `mapstructure:"INBOX_MAX_RETRY_ATTEMPTS"`
	RetryBackoff     time.Duration `mapstructure:"INBOX_RETRY_BACKOFF"`
}

// initInboxProcessorConfig initializes the inbox processor configuration.
func initInboxProcessorConfig() *InboxProcessorConfig {
	return &InboxProcessorConfig{
		PollInterval:     time.Duration(viper.GetInt("INBOX_POLL_INTERVAL")) * time.Second,
		CleanupInterval:  time.Duration(viper.GetInt("INBOX_CLEANUP_INTERVAL")) * time.Hour,
		RetentionPeriod:  time.Duration(viper.GetInt("INBOX_RETENTION_PERIOD")) * time.Hour,
		BatchSize:        viper.GetInt64("INBOX_BATCH_SIZE"),
		MaxRetryAttempts: viper.GetInt64("INBOX_MAX_RETRY_ATTEMPTS"),
		RetryBackoff:     time.Duration(viper.GetInt("INBOX_RETRY_BACKOFF")) * time.Second,
	}
}
