package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultSagaExecutionTimeout  = 30 * time.Minute
	defaultSagaMaxConcurrent     = 100
	defaultSagaDefaultMaxRetries = 3
	defaultSagaDefaultRetryDelay = 2 * time.Second
	defaultSagaMaxRetryDelay     = 1 * time.Minute
	defaultSagaRecoveryInterval  = 5 * time.Minute
	defaultSagaRecoveryBatchSize = 100
	defaultSagaMaxRecoveryAge    = 24 * time.Hour
	defaultSagaStateRetention    = 30 * 24 * time.Hour
	defaultSagaPurgeInterval     = 24 * time.Hour
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
	// Set defaults
	viper.SetDefault("SAGA_EXECUTION_MODE", "async")
	viper.SetDefault("SAGA_DEFAULT_EXECUTION_TIMEOUT", defaultSagaExecutionTimeout)
	viper.SetDefault("SAGA_MAX_CONCURRENT", defaultSagaMaxConcurrent)
	viper.SetDefault("SAGA_DEFAULT_MAX_RETRIES", defaultSagaDefaultMaxRetries)
	viper.SetDefault("SAGA_DEFAULT_RETRY_DELAY", defaultSagaDefaultRetryDelay)
	viper.SetDefault("SAGA_EXPONENTIAL_BACKOFF", true)
	viper.SetDefault("SAGA_MAX_RETRY_DELAY", defaultSagaMaxRetryDelay)
	viper.SetDefault("SAGA_RECOVERY_ENABLED", true)
	viper.SetDefault("SAGA_RECOVERY_INTERVAL", defaultSagaRecoveryInterval)
	viper.SetDefault("SAGA_RECOVERY_BATCH_SIZE", defaultSagaRecoveryBatchSize)
	viper.SetDefault("SAGA_MAX_RECOVERY_AGE", defaultSagaMaxRecoveryAge)
	viper.SetDefault("SAGA_STATE_RETENTION", defaultSagaStateRetention)
	viper.SetDefault("SAGA_PURGE_COMPLETED", true)
	viper.SetDefault("SAGA_PURGE_INTERVAL", defaultSagaPurgeInterval)

	sagaConfig := &SagaConfig{}
	if err := viper.Unmarshal(&sagaConfig); err != nil {
		panic(err)
	}

	return sagaConfig
}
