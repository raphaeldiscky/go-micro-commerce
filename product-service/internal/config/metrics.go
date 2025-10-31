package config

import (
	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// MetricsConfig holds configuration for Prometheus metrics.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"METRICS_ENABLED"`
	Path    string `mapstructure:"METRICS_PATH"`
}

// initMetricsConfig initializes the metrics configuration from environment variables.
func initMetricsConfig() *MetricsConfig {
	viper.SetDefault("METRICS_ENABLED", constant.MetricsEnabled)
	viper.SetDefault("METRICS_PATH", constant.MetricsPath)

	metricsConfig := &MetricsConfig{}
	if err := viper.Unmarshal(metricsConfig); err != nil {
		panic(err)
	}

	return metricsConfig
}
