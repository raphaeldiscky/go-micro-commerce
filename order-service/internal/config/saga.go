package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// SagaConfig holds saga configuration.
type SagaConfig struct {
	// Execution settings
	ExecutionMode           string        `mapstructure:"SAGA_EXECUTION_MODE"` // async or sync
	DefaultExecutionTimeout time.Duration `mapstructure:"SAGA_DEFAULT_EXECUTION_TIMEOUT"`
	MaxConcurrentSagas      int           `mapstructure:"SAGA_MAX_CONCURRENT"`

	// Retry settings
	DefaultMaxRetries  int           `mapstructure:"SAGA_DEFAULT_MAX_RETRIES"`
	DefaultRetryDelay  time.Duration `mapstructure:"SAGA_DEFAULT_RETRY_DELAY"`
	ExponentialBackoff bool          `mapstructure:"SAGA_EXPONENTIAL_BACKOFF"`
	MaxRetryDelay      time.Duration `mapstructure:"SAGA_MAX_RETRY_DELAY"`

	// Recovery settings
	RecoveryEnabled   bool          `mapstructure:"SAGA_RECOVERY_ENABLED"`
	RecoveryInterval  time.Duration `mapstructure:"SAGA_RECOVERY_INTERVAL"`
	RecoveryBatchSize int           `mapstructure:"SAGA_RECOVERY_BATCH_SIZE"`
	MaxRecoveryAge    time.Duration `mapstructure:"SAGA_MAX_RECOVERY_AGE"`

	// Persistence settings
	StateRetentionPeriod time.Duration `mapstructure:"SAGA_STATE_RETENTION"`
	PurgeCompletedSagas  bool          `mapstructure:"SAGA_PURGE_COMPLETED"`
	PurgeInterval        time.Duration `mapstructure:"SAGA_PURGE_INTERVAL"`
}

// initSagaConfig initializes the saga configuration from environment variables.
func initSagaConfig() *SagaConfig {
	viper.SetDefault("SAGA_EXECUTION_MODE", "async")
	viper.SetDefault("SAGA_DEFAULT_EXECUTION_TIMEOUT", constant.SagaDefaultExecutionTimeout)
	viper.SetDefault("SAGA_MAX_CONCURRENT", constant.SagaMaxConcurrent)
	viper.SetDefault("SAGA_DEFAULT_MAX_RETRIES", constant.SagaDefaultMaxRetries)
	viper.SetDefault("SAGA_DEFAULT_RETRY_DELAY", constant.SagaDefaultRetryDelay)
	viper.SetDefault("SAGA_EXPONENTIAL_BACKOFF", true)
	viper.SetDefault("SAGA_MAX_RETRY_DELAY", constant.SagaMaxRetryDelay)
	viper.SetDefault("SAGA_RECOVERY_ENABLED", true)
	viper.SetDefault("SAGA_RECOVERY_INTERVAL", constant.SagaRecoveryInterval)
	viper.SetDefault("SAGA_RECOVERY_BATCH_SIZE", constant.SagaRecoveryBatchSize)
	viper.SetDefault("SAGA_MAX_RECOVERY_AGE", constant.SagaMaxRecoveryAge)
	viper.SetDefault("SAGA_STATE_RETENTION", constant.SagaStateRetention)
	viper.SetDefault("SAGA_PURGE_COMPLETED", true)
	viper.SetDefault("SAGA_PURGE_INTERVAL", constant.SagaPurgeInterval)

	sagaConfig := &SagaConfig{}
	if err := viper.Unmarshal(&sagaConfig); err != nil {
		panic(err)
	}

	return sagaConfig
}
