package config

import (
	"log"

	"github.com/spf13/viper"
)

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"METRICS_ENABLED"`
	Path    string `mapstructure:"METRICS_PATH"`
	Port    int    `mapstructure:"METRICS_PORT"`
}

// initMetricsConfig initializes the metrics configuration from environment variables.
func initMetricsConfig() *MetricsConfig {
	viper.SetDefault("METRICS_ENABLED", true)
	viper.SetDefault("METRICS_PATH", "/metrics")
	viper.SetDefault("METRICS_PORT", 9090)

	metricsConfig := &MetricsConfig{}

	if err := viper.Unmarshal(&metricsConfig); err != nil {
		log.Fatalf("error mapping metrics config: %v", err)
	}

	return metricsConfig
}
