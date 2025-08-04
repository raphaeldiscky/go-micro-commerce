package config

import (
	"log"

	"github.com/spf13/viper"
)

// TracingConfig holds tracing configuration.
type TracingConfig struct {
	Enabled       bool    `mapstructure:"TRACING_ENABLED"`
	JaegerURL     string  `mapstructure:"TRACING_JAEGER_URL"`
	ServiceName   string  `mapstructure:"TRACING_SERVICE_NAME"`
	SamplingRate  float64 `mapstructure:"TRACING_SAMPLING_RATE"`
	Environment   string  `mapstructure:"TRACING_ENVIRONMENT"`
	BatchTimeout  int     `mapstructure:"TRACING_BATCH_TIMEOUT"`
	ExportTimeout int     `mapstructure:"TRACING_EXPORT_TIMEOUT"`
}

// initTracingConfig initializes the tracing configuration from environment variables.
func initTracingConfig() *TracingConfig {
	tracingConfig := &TracingConfig{}

	if err := viper.Unmarshal(&tracingConfig); err != nil {
		log.Fatalf("error mapping tracing config: %v", err)
	}

	return tracingConfig
}
