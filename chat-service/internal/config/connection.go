// Package config provides configuration management for the application.
package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// ConnectionConfig holds the connection service configuration.
type ConnectionConfig struct {
	PublicKeyPath       string        `mapstructure:"CONN_PUBLIC_KEY_PATH"`
	JWKSUrl             string        `mapstructure:"CONN_JWKS_URL"`
	JWKSCacheTTL        time.Duration `mapstructure:"CONN_JWKS_CACHE_TTL"`
	JWKSRefreshInterval time.Duration `mapstructure:"CONN_JWKS_REFRESH_INTERVAL"`
	DefaultNodeAddress  string        `mapstructure:"CONN_DEFAULT_NODE_ADDRESS"`
	MaxConnections      int           `mapstructure:"CONN_MAX_CONNECTIONS"`
	ConsulAddress       string        `mapstructure:"CONN_CONSUL_ADDRESS"`
	ChatServiceName     string        `mapstructure:"CONN_CHAT_SERVICE_NAME"`
}

// initConnectionConfig initializes the connection configuration from environment variables.
func initConnectionConfig() *ConnectionConfig {
	// Explicitly bind JWKS environment variables
	if err := viper.BindEnv("CONN_JWKS_URL"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("CONN_PUBLIC_KEY_PATH"); err != nil {
		panic(err)
	}

	// Set defaults for non-critical fields
	viper.SetDefault("CONN_JWKS_CACHE_TTL", "1h")
	viper.SetDefault("CONN_JWKS_REFRESH_INTERVAL", "15m")
	viper.SetDefault("CONN_DEFAULT_NODE_ADDRESS", "ws://localhost:9098")
	viper.SetDefault("CONN_MAX_CONNECTIONS", constant.ConnMaxConnections)
	viper.SetDefault("CONN_CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("CONN_CHAT_SERVICE_NAME", constant.AppName+"-ws")

	connectionConfig := &ConnectionConfig{}
	if err := viper.Unmarshal(connectionConfig); err != nil {
		panic(err)
	}

	return connectionConfig
}
