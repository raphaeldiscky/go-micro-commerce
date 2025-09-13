package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/constant"
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
	viper.SetDefault("ELASTICSEARCH_PORT", constant.ElasticPort)
	viper.SetDefault("ELASTICSEARCH_USERNAME", "elastic")
	viper.SetDefault("ELASTICSEARCH_PASSWORD", "elasticsearch")
	viper.SetDefault("ELASTICSEARCH_ENABLE_SECURITY", true)
	viper.SetDefault("ELASTICSEARCH_ENABLE_SSL", false)
	viper.SetDefault("ELASTICSEARCH_MAX_RETRIES", constant.ElasticMaxRetries)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_CONNS", constant.ElasticMaxIdleConns)
	viper.SetDefault("ELASTICSEARCH_MAX_IDLE_TIME", constant.ElasticMaxIdleTime)
	viper.SetDefault("ELASTICSEARCH_REQUEST_TIMEOUT", constant.ElasticRequestTimeout)
	viper.SetDefault("ELASTICSEARCH_DISCOVER_NODES_INTERVAL", constant.ElasticDiscoverNodesInterval)
	viper.SetDefault("ELASTICSEARCH_SNIFFER_ENABLED", false)
	viper.SetDefault("ELASTICSEARCH_HEALTHCHECK_URL", "/_cluster/health")

	esConfig := &ElasticsearchConfig{}
	if err := viper.Unmarshal(esConfig); err != nil {
		panic(err)
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
