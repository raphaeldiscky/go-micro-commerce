package config

import (
	"github.com/spf13/viper"
)

const (
	defaultMetricsPort = 9090
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
	viper.SetDefault("METRICS_PORT", defaultMetricsPort)

	metricsConfig := &MetricsConfig{}
	if err := viper.Unmarshal(metricsConfig); err != nil {
		panic(err)
	}

	return metricsConfig
}
