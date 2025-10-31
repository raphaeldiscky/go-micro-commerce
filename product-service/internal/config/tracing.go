package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// TracingConfig holds configuration for OpenTelemetry tracing.
type TracingConfig struct {
	Enabled       bool          `mapstructure:"TRACING_ENABLED"`
	URL           string        `mapstructure:"TRACING_URL"`
	ServiceName   string        `mapstructure:"TRACING_SERVICE_NAME"`
	SamplingRate  float64       `mapstructure:"TRACING_SAMPLING_RATE"`
	Environment   string        `mapstructure:"TRACING_ENVIRONMENT"`
	BatchTimeout  time.Duration `mapstructure:"TRACING_BATCH_TIMEOUT"`
	ExportTimeout time.Duration `mapstructure:"TRACING_EXPORT_TIMEOUT"`
}

// initTracingConfig initializes the tracing configuration from environment variables.
func initTracingConfig() *TracingConfig {
	viper.SetDefault("TRACING_ENABLED", constant.TracingEnabled)
	viper.SetDefault("TRACING_URL", constant.TracingURL)
	viper.SetDefault("TRACING_SERVICE_NAME", constant.TracingServiceName)
	viper.SetDefault("TRACING_SAMPLING_RATE", constant.TracingSamplingRate)
	viper.SetDefault("TRACING_ENVIRONMENT", constant.TracingEnvironment)
	viper.SetDefault("TRACING_BATCH_TIMEOUT", constant.TracingBatchTimeout)
	viper.SetDefault("TRACING_EXPORT_TIMEOUT", constant.TracingExportTimeout)

	tracingConfig := &TracingConfig{}
	if err := viper.Unmarshal(tracingConfig); err != nil {
		panic(err)
	}

	return tracingConfig
}
