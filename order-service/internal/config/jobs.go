package config

import (
	"time"

	"github.com/spf13/viper"
)

// JobConfig holds the job configuration.
type JobConfig struct {
	Recovery *JobRecoveryConfig
	// Add other job configs here as needed
	// Cleanup      *CleanupJobConfig
}

// JobRecoveryConfig holds the  recovery job configuration.
type JobRecoveryConfig struct {
	Enabled             bool
	Interval            time.Duration
	MaxRetries          int64
	MaxRowsFetch        int64
	Timeout             time.Duration
	MaxAge              time.Duration
	RedisLockTTL        time.Duration
	RedisLockBackoff    time.Duration
	RedisLockMaxRetries int
}

const (
	defaultJobRecoveryInterval     = 5 * time.Minute
	defaultJobRecoveryMaxRetries   = 5
	defaultJobRecoveryMaxAge       = 24 * time.Hour
	defaultJobRecoveryTimeout      = 30 * time.Second
	defaultJobRecoveryMaxRowsFetch = 100
	jobRecoveryRedisLockTTL        = 10 * time.Minute
	jobRecoveryRedisLockBackoff    = 100 * time.Millisecond
	jobRecoveryRedisLockMaxRetries = 10
)

// initJobConfig initializes the job configuration from environment variables.
func initJobConfig() *JobConfig {
	setJobDefaults()

	return &JobConfig{
		Recovery: &JobRecoveryConfig{
			Enabled:             viper.GetBool("JOB_RECOVERY_ENABLED"),
			Interval:            viper.GetDuration("JOB_RECOVERY_INTERVAL"),
			MaxRetries:          viper.GetInt64("JOB_RECOVERY_MAX_RETRIES"),
			MaxAge:              viper.GetDuration("JOB_RECOVERY_MAX_AGE"),
			Timeout:             viper.GetDuration("JOB_RECOVERY_TIMEOUT"),
			MaxRowsFetch:        viper.GetInt64("JOB_RECOVERY_MAX_ROWS_FETCH"),
			RedisLockTTL:        viper.GetDuration("JOB_RECOVERY_REDIS_LOCK_TTL"),
			RedisLockBackoff:    viper.GetDuration("JOB_RECOVERY_REDIS_LOCK_BACKOFF"),
			RedisLockMaxRetries: viper.GetInt("JOB_RECOVERY_REDIS_LOCK_MAX_RETRIES"),
		},
		// Add other job configs here
		// Cleanup: &JobCleanupConfig{
		//     Enabled:  viper.GetBool("JOB_CLEANUP_ENABLED"),
		//     Interval: viper.GetDuration("JOB_CLEANUP_INTERVAL"),
		// },
	}
}

// setJobDefaults sets default values for job configuration.
func setJobDefaults() {
	//  Recovery Job defaults
	viper.SetDefault("JOB_RECOVERY_ENABLED", true)
	viper.SetDefault("JOB_RECOVERY_INTERVAL", defaultJobRecoveryInterval)
	viper.SetDefault("JOB_RECOVERY_MAX_RETRIES", defaultJobRecoveryMaxRetries)
	viper.SetDefault("JOB_RECOVERY_MAX_AGE", defaultJobRecoveryMaxAge)
	viper.SetDefault("JOB_RECOVERY_TIMEOUT", defaultJobRecoveryTimeout)
	viper.SetDefault("JOB_RECOVERY_MAX_ROWS_FETCH", defaultJobRecoveryMaxRowsFetch)

	viper.SetDefault("JOB_RECOVERY_REDIS_LOCK_TTL", jobRecoveryRedisLockTTL)
	viper.SetDefault("JOB_RECOVERY_REDIS_LOCK_BACKOFF", jobRecoveryRedisLockBackoff)
	viper.SetDefault("JOB_RECOVERY_REDIS_LOCK_MAX_RETRIES", jobRecoveryRedisLockMaxRetries)
	// Add other job defaults here
	// viper.SetDefault("JOB_CLEANUP_ENABLED", true)
	// viper.SetDefault("JOB_CLEANUP_INTERVAL", "1h")
}
