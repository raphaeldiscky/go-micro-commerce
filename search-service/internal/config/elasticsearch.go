package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultElasticPort                                = 9200
	defaultElasticMaxRetries                          = 3
	defaultElasticMaxIdleConns                        = 10 * 1024
	defaultElasticMaxIdleTime                         = 30 * time.Second
	defaultElasticRequestTimeout                      = 30 * time.Second
	defaultElasticDiscoverNodesInterval time.Duration = 60 * time.Second
)

// ElasticsearchConfig holds the configuration for Elasticsearch connection.
type ElasticsearchConfig struct {
	Host                  string        `mapstructure:"ELASTICSEARCH_HOST"`
	Port                  int           `mapstructure:"ELASTICSEARCH_PORT"`
	Username              string        `mapstructure:"ELASTICSEARCH_USERNAME"`
	Password              string        `mapstructure:"ELASTICSEARCH_PASSWORD"`
	EnableSecurity        bool          `mapstructure:"ELASTICSEARCH_ENABLE_SECURITY"`
	EnableSSL             bool          `mapstructure:"ELASTICSEARCH_ENABLE_SSL"`
	MaxRetries            int           `mapstructure:"ELASTICSEARCH_MAX_RETRIES"`
	MaxIdleConns          int           `mapstructure:"ELASTICSEARCH_MAX_IDLE_CONNS"`
	MaxIdleTime           time.Duration `mapstructure:"ELASTICSEARCH_MAX_IDLE_TIME"`
	RequestTimeout        time.Duration `mapstructure:"ELASTICSEARCH_REQUEST_TIMEOUT"`
	DiscoverNodesInterval time.Duration `mapstructure:"ELASTICSEARCH_DISCOVER_NODES_INTERVAL"`
	SnifferEnabled        bool          `mapstructure:"ELASTICSEARCH_SNIFFER_ENABLED"`
	HealthcheckURL        string        `mapstructure:"ELASTICSEARCH_HEALTHCHECK_URL"`
}

// initElasticsearchConfig initializes Elasticsearch configuration with defaults.
func initElasticsearchConfig() *ElasticsearchConfig {
	viper.SetDefault("ELASTICSEARCH_HOST", "localhost")
	viper.SetDefault("ELASTICSEARCH_PORT", defaultElasticPort)
	viper.SetDefault("ELASTICSEARCH_USERNAME", "elastic")
	viper.SetDefault("ELASTICSEARCH_PASSWORD", "elasticsearch")
	viper.SetDefault("ELASTICSEARCH_ENABLE_SECURITY", true)
	viper.SetDefault("ELASTICSEARCH_ENABLE_SSL", false)
	viper.SetDefault("ELASTICSEARCH_MAX_RETRIES", defaultElasticMaxRetries)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_CONNS", defaultElasticMaxIdleConns)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_TIME", defaultElasticMaxIdleTime)
	viper.SetDefault("ELASTICSEARCH_REQUEST_TIMEOUT", defaultElasticRequestTimeout)
	viper.SetDefault("ELASTICSEARCH_DISCOVER_NODES_INTERVAL", defaultElasticDiscoverNodesInterval)
	viper.SetDefault("ELASTICSEARCH_SNIFFER_ENABLED", false)
	viper.SetDefault("ELASTICSEARCH_HEALTHCHECK_URL", "/_cluster/health")

	esConfig := &ElasticsearchConfig{}

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
