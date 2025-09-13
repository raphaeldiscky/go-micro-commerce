package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/constant"
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
	viper.SetDefault("INBOX_POLL_INTERVAL", constant.InboxPollInterval)
	viper.SetDefault("INBOX_CLEANUP_INTERVAL", constant.InboxCleanupInterval)
	viper.SetDefault("INBOX_RETENTION_PERIOD", constant.InboxRetentionPeriod)
	viper.SetDefault("INBOX_BATCH_SIZE", constant.InboxBatchSize)
	viper.SetDefault("INBOX_MAX_RETRY_ATTEMPTS", constant.InboxMaxRetryAttempts)
	viper.SetDefault("INBOX_RETRY_BACKOFF", constant.InboxRetryBackoff)

	inboxProcessorConfig := &InboxProcessorConfig{}

	if err := viper.Unmarshal(inboxProcessorConfig); err != nil {
		panic(err)
	}

	return inboxProcessorConfig
}
