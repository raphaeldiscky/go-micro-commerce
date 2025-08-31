package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// SagaConfig holds saga configuration.
type SagaConfig struct {
	// Execution settings
	ExecutionMode      string        `mapstructure:"SAGA_EXECUTION_MODE"` // async or sync
	AsyncTimeout       time.Duration `mapstructure:"SAGA_ASYNC_TIMEOUT"`
	MaxConcurrentSagas int           `mapstructure:"SAGA_MAX_CONCURRENT"`

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
	viper.SetDefault("SAGA_ASYNC_TIMEOUT", 30*time.Minute)
	viper.SetDefault("SAGA_MAX_CONCURRENT", 100)
	viper.SetDefault("SAGA_DEFAULT_MAX_RETRIES", 3)
	viper.SetDefault("SAGA_DEFAULT_RETRY_DELAY", 2*time.Second)
	viper.SetDefault("SAGA_EXPONENTIAL_BACKOFF", true)
	viper.SetDefault("SAGA_MAX_RETRY_DELAY", 1*time.Minute)
	viper.SetDefault("SAGA_RECOVERY_ENABLED", true)
	viper.SetDefault("SAGA_RECOVERY_INTERVAL", 5*time.Minute)
	viper.SetDefault("SAGA_RECOVERY_BATCH_SIZE", 100)
	viper.SetDefault("SAGA_MAX_RECOVERY_AGE", 24*time.Hour)
	viper.SetDefault("SAGA_STATE_RETENTION", 30*24*time.Hour)
	viper.SetDefault("SAGA_PURGE_COMPLETED", true)
	viper.SetDefault("SAGA_PURGE_INTERVAL", 24*time.Hour)

	sagaConfig := &SagaConfig{}
	if err := viper.Unmarshal(&sagaConfig); err != nil {
		log.Fatalf("error mapping Saga config: %v", err)
	}

	return sagaConfig
}
