package config

import (
	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/constant"
)

// TracingConfig holds tracing configuration.
type TracingConfig struct {
	Enabled       bool    `mapstructure:"TRACING_ENABLED"`
	URL           string  `mapstructure:"TRACING_URL"`
	ServiceName   string  `mapstructure:"TRACING_SERVICE_NAME"`
	SamplingRate  float64 `mapstructure:"TRACING_SAMPLING_RATE"`
	Environment   string  `mapstructure:"TRACING_ENVIRONMENT"`
	BatchTimeout  int     `mapstructure:"TRACING_BATCH_TIMEOUT"`
	ExportTimeout int     `mapstructure:"TRACING_EXPORT_TIMEOUT"`
}

// initTracingConfig initializes the tracing configuration from environment variables.
func initTracingConfig() *TracingConfig {
	viper.SetDefault("TRACING_ENABLED", true)
	viper.SetDefault("TRACING_URL", "http://localhost:4318/v1/traces")
	viper.SetDefault("TRACING_SERVICE_NAME", "api-gateway")
	viper.SetDefault("TRACING_SAMPLING_RATE", constant.TracingSamplingRate)
	viper.SetDefault("TRACING_ENVIRONMENT", "development")
	viper.SetDefault("TRACING_BATCH_TIMEOUT", constant.TracingBatchTimeout)
	viper.SetDefault("TRACING_EXPORT_TIMEOUT", constant.TracingExportTimeout)

	tracingConfig := &TracingConfig{}
	if err := viper.Unmarshal(tracingConfig); err != nil {
		panic(err)
	}

	return tracingConfig
}
