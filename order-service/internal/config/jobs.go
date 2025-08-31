package config

import (
	"time"

	"github.com/spf13/viper"
)

// JobsConfig holds the jobs configuration.
type JobsConfig struct {
	SagaRecovery *SagaRecoveryJobConfig
	// Add other job configs here as needed
	// Cleanup      *CleanupJobConfig
}

// SagaRecoveryJobConfig holds the saga recovery job configuration.
type SagaRecoveryJobConfig struct {
	Enabled    bool
	Interval   time.Duration
	MaxRetries int64
	MaxAge     time.Duration
}

// initJobsConfig initializes the jobs configuration from environment variables.
func initJobsConfig() *JobsConfig {
	return &JobsConfig{
		SagaRecovery: &SagaRecoveryJobConfig{
			Enabled:    viper.GetBool("JOBS_SAGA_RECOVERY_ENABLED"),
			Interval:   viper.GetDuration("JOBS_SAGA_RECOVERY_INTERVAL"),
			MaxRetries: viper.GetInt64("JOBS_SAGA_RECOVERY_MAX_RETRIES"),
			MaxAge:     viper.GetDuration("JOBS_SAGA_RECOVERY_MAX_AGE"),
		},
		// Add other job configs here
		// Cleanup: &CleanupJobConfig{
		//     Enabled:  viper.GetBool("JOBS_CLEANUP_ENABLED"),
		//     Interval: viper.GetDuration("JOBS_CLEANUP_INTERVAL"),
		// },
	}
}

// setJobsDefaults sets default values for jobs configuration.
func setJobsDefaults() {
	// Saga Recovery Job defaults
	viper.SetDefault("JOBS_SAGA_RECOVERY_ENABLED", true)
	viper.SetDefault("JOBS_SAGA_RECOVERY_INTERVAL", "5m")
	viper.SetDefault("JOBS_SAGA_RECOVERY_MAX_RETRIES", 5)
	viper.SetDefault("JOBS_SAGA_RECOVERY_MAX_AGE", "24h")
	// Add other job defaults here
	// viper.SetDefault("JOBS_CLEANUP_ENABLED", true)
	// viper.SetDefault("JOBS_CLEANUP_INTERVAL", "1h")
}
