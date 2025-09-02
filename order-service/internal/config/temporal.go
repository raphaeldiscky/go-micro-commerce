package config

import (
	"time"

	"github.com/spf13/viper"
)

// TemporalConfig holds the Temporal configuration.
type TemporalConfig struct {
	HostPort        string        `mapstructure:"TEMPORAL_HOST_PORT"`
	Namespace       string        `mapstructure:"TEMPORAL_NAMESPACE"`
	TaskQueue       string        `mapstructure:"TEMPORAL_TASK_QUEUE"`
	WorkerTimeout   time.Duration `mapstructure:"TEMPORAL_WORKER_TIMEOUT"`
	WorkflowTimeout time.Duration `mapstructure:"TEMPORAL_WORKFLOW_TIMEOUT"`
	ActivityTimeout time.Duration `mapstructure:"TEMPORAL_ACTIVITY_TIMEOUT"`
}

// initTemporalConfig initializes the Temporal configuration.
func initTemporalConfig() *TemporalConfig {
	setTemporalDefaults()

	return &TemporalConfig{
		HostPort:        viper.GetString("TEMPORAL_HOST_PORT"),
		Namespace:       viper.GetString("TEMPORAL_NAMESPACE"),
		TaskQueue:       viper.GetString("TEMPORAL_TASK_QUEUE"),
		WorkerTimeout:   viper.GetDuration("TEMPORAL_WORKER_TIMEOUT"),
		WorkflowTimeout: viper.GetDuration("TEMPORAL_WORKFLOW_TIMEOUT"),
		ActivityTimeout: viper.GetDuration("TEMPORAL_ACTIVITY_TIMEOUT"),
	}
}

// setTemporalDefaults sets default values for Temporal configuration.
func setTemporalDefaults() {
	viper.SetDefault("TEMPORAL_HOST_PORT", "localhost:7233")
	viper.SetDefault("TEMPORAL_NAMESPACE", "default")
	viper.SetDefault("TEMPORAL_TASK_QUEUE", "order-saga-task-queue")
	viper.SetDefault("TEMPORAL_WORKER_TIMEOUT", 30*time.Second)
	viper.SetDefault("TEMPORAL_WORKFLOW_TIMEOUT", 30*time.Minute)
	viper.SetDefault("TEMPORAL_ACTIVITY_TIMEOUT", 5*time.Minute)
}
