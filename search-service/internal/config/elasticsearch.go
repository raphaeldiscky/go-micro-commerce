package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ElasticsearchConfig holds the configuration for Elasticsearch connection.
type ElasticsearchConfig struct {
	Host           string `mapstructure:"ELASTICSEARCH_HOST"`
	Port           int    `mapstructure:"ELASTICSEARCH_PORT"`
	Username       string `mapstructure:"ELASTICSEARCH_USERNAME"`
	Password       string `mapstructure:"ELASTICSEARCH_PASSWORD"`
	EnableSecurity bool   `mapstructure:"ELASTICSEARCH_ENABLE_SECURITY"`
	EnableSSL      bool   `mapstructure:"ELASTICSEARCH_ENABLE_SSL"`
	MaxRetries     int    `mapstructure:"ELASTICSEARCH_MAX_RETRIES"`
	MaxIdleConns   int    `mapstructure:"ELASTICSEARCH_MAX_IDLE_CONNS"`
	MaxIdleTime    int    `mapstructure:"ELASTICSEARCH_MAX_IDLE_TIME"`
	RequestTimeout int    `mapstructure:"ELASTICSEARCH_REQUEST_TIMEOUT"`
	SnifferEnabled bool   `mapstructure:"ELASTICSEARCH_SNIFFER_ENABLED"`
	HealthcheckURL string `mapstructure:"ELASTICSEARCH_HEALTHCHECK_URL"`
}

// initElasticsearchConfig initializes Elasticsearch configuration with defaults.
func initElasticsearchConfig() *ElasticsearchConfig {
	// Set defaults for Elasticsearch v9
	viper.SetDefault("ELASTICSEARCH_HOST", "localhost")
	viper.SetDefault("ELASTICSEARCH_PORT", 9200)
	viper.SetDefault("ELASTICSEARCH_USERNAME", "elastic")
	viper.SetDefault("ELASTICSEARCH_PASSWORD", "elasticsearch")
	viper.SetDefault("ELASTICSEARCH_ENABLE_SECURITY", true)
	viper.SetDefault("ELASTICSEARCH_ENABLE_SSL", false)
	viper.SetDefault("ELASTICSEARCH_MAX_RETRIES", 3)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_CONNS", 10)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_TIME", 30)
	viper.SetDefault("ELASTICSEARCH_REQUEST_TIMEOUT", 30)
	viper.SetDefault("ELASTICSEARCH_SNIFFER_ENABLED", false)
	viper.SetDefault("ELASTICSEARCH_HEALTHCHECK_URL", "/_cluster/health")

	esConfig := &ElasticsearchConfig{}

	if err := viper.Unmarshal(&esConfig); err != nil {
		log.Fatalf("error mapping elasticsearch config: %v", err)
	}

	return esConfig
}

// GetElasticsearchURL returns the full Elasticsearch URL.
func (c *ElasticsearchConfig) GetElasticsearchURL() string {
	protocol := "http"
	if c.EnableSSL {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
}
