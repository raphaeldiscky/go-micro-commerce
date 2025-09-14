package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// TemporalConfig holds the Temporal configuration.
type TemporalConfig struct {
	APIKey                      string        `mapstructure:"TEMPORAL_API_KEY"`
	Address                     string        `mapstructure:"TEMPORAL_ADDRESS"`
	Namespace                   string        `mapstructure:"TEMPORAL_NAMESPACE"`
	TaskQueue                   string        `mapstructure:"TEMPORAL_TASK_QUEUE"`
	RetryInterval               time.Duration `mapstructure:"TEMPORAL_RETRY_INTERVAL"`
	BackoffCoefficient          float64       `mapstructure:"TEMPORAL_BACKOFF_COEFFICIENT"`
	MaxAttempts                 int32         `mapstructure:"TEMPORAL_MAX_ATTEMPTS"`
	MaxInterval                 time.Duration `mapstructure:"TEMPORAL_MAX_INTERVAL"`
	WorkflowTimeout             time.Duration `mapstructure:"TEMPORAL_WORKFLOW_TIMEOUT"`
	CompensationWorkflowTimeout time.Duration `mapstructure:"TEMPORAL_COMPENSATION_WORKFLOW_TIMEOUT"`
}

// initTemporalConfig initializes the Temporal configuration.
func initTemporalConfig() *TemporalConfig {
	viper.SetDefault("TEMPORAL_API_KEY", "supersecret")
	viper.SetDefault("TEMPORAL_ADDRESS", "localhost:7233")
	viper.SetDefault("TEMPORAL_NAMESPACE", "default")
	viper.SetDefault("TEMPORAL_TASK_QUEUE", "order-saga-task-queue")
	viper.SetDefault("TEMPORAL_WORKFLOW_TIMEOUT", constant.TemporalWorkflowTimeout)
	viper.SetDefault(
		"TEMPORAL_COMPENSATION_WORKFLOW_TIMEOUT",
		constant.TemporalCompensationWorkflowTimeout,
	)
	viper.SetDefault("TEMPORAL_RETRY_INTERVAL", constant.TemporalRetryInterval)
	viper.SetDefault("TEMPORAL_BACKOFF_COEFFICIENT", constant.TemporalBackoffCoefficient)
	viper.SetDefault("TEMPORAL_MAX_ATTEMPTS", constant.TemporalMaxAttempts)
	viper.SetDefault("TEMPORAL_MAX_INTERVAL", constant.TemporalMaxInterval)

	temporalConfig := &TemporalConfig{}
	if err := viper.Unmarshal(&temporalConfig); err != nil {
		panic(err)
	}

	return temporalConfig
}
