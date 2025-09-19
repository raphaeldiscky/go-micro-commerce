package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/constant"
)

// ESConfig holds the configuration for ES connection.
type ESConfig struct {
	Host                  string        `mapstructure:"ES_HOST"`
	Port                  int           `mapstructure:"ES_PORT"`
	Username              string        `mapstructure:"ES_USERNAME"`
	Password              string        `mapstructure:"ES_PASSWORD"`
	EnableSecurity        bool          `mapstructure:"ES_ENABLE_SECURITY"`
	EnableSSL             bool          `mapstructure:"ES_ENABLE_SSL"`
	SkipTLSVerify         bool          `mapstructure:"ES_SKIP_TLS_VERIFY"`
	MaxRetries            int           `mapstructure:"ES_MAX_RETRIES"`
	MaxIdleConns          int           `mapstructure:"ES_MAX_IDLE_CONNS"`
	MaxIdleTime           time.Duration `mapstructure:"ES_MAX_IDLE_TIME"`
	RequestTimeout        time.Duration `mapstructure:"ES_REQUEST_TIMEOUT"`
	DiscoverNodesInterval time.Duration `mapstructure:"ES_DISCOVER_NODES_INTERVAL"`
	SnifferEnabled        bool          `mapstructure:"ES_SNIFFER_ENABLED"`
	HealthcheckURL        string        `mapstructure:"ES_HEALTHCHECK_URL"`
}

// initESConfig initializes ES configuration with defaults.
func initESConfig() *ESConfig {
	viper.SetDefault("ES_HOST", "localhost")
	viper.SetDefault("ES_PORT", constant.ElasticPort)
	viper.SetDefault("ES_USERNAME", "elastic")
	viper.SetDefault("ES_PASSWORD", "elastic")
	viper.SetDefault("ES_ENABLE_SECURITY", true)
	viper.SetDefault("ES_ENABLE_SSL", true)
	viper.SetDefault("ES_SKIP_TLS_VERIFY", false)
	viper.SetDefault("ES_MAX_RETRIES", constant.ElasticMaxRetries)
	viper.SetDefault("ES_MAX_IDLE_CONNS", constant.ElasticMaxIdleConns)
	viper.SetDefault("ES_MAX_IDLE_TIME", constant.ElasticMaxIdleTime)
	viper.SetDefault("ES_REQUEST_TIMEOUT", constant.ElasticRequestTimeout)
	viper.SetDefault("ES_DISCOVER_NODES_INTERVAL", constant.ElasticDiscoverNodesInterval)
	viper.SetDefault("ES_SNIFFER_ENABLED", false)
	viper.SetDefault("ES_HEALTHCHECK_URL", "/_cluster/health")

	esConfig := &ESConfig{}
	if err := viper.Unmarshal(esConfig); err != nil {
		panic(err)
	}

	return esConfig
}

// GetESURL returns the full ES URL.
func (c *ESConfig) GetESURL() string {
	protocol := "http"
	if c.EnableSSL {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
}
