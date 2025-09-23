// Package config provides configuration management for the application.
package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// ConnectionConfig holds the connection service configuration.
type ConnectionConfig struct {
	JWTSecret          string        `mapstructure:"CONN_JWT_SECRET"`
	TicketExpiration   time.Duration `mapstructure:"CONN_TICKET_EXPIRATION"`
	DefaultNodeAddress string        `mapstructure:"CONN_DEFAULT_NODE_ADDRESS"`
	MaxConnections     int           `mapstructure:"CONN_MAX_CONNECTIONS"`
	ConsulAddress      string        `mapstructure:"CONN_CONSUL_ADDRESS"`
	ChatServiceName    string        `mapstructure:"CONN_CHAT_SERVICE_NAME"`
}

// initConnectionConfig initializes the connection configuration from environment variables.
func initConnectionConfig() *ConnectionConfig {
	// Set defaults
	viper.SetDefault("CONN_JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("CONN_TICKET_EXPIRATION", constant.ConnMaxConnections)
	viper.SetDefault("CONN_DEFAULT_NODE_ADDRESS", "ws://localhost:9088")
	viper.SetDefault("CONN_MAX_CONNECTIONS", constant.ConnMaxConnections)
	viper.SetDefault("CONN_CONSUL_ADDRESS", "localhost:8500")
	viper.SetDefault("CONN_CHAT_SERVICE_NAME", "chat-service-websocket")

	connectionConfig := &ConnectionConfig{}
	if err := viper.Unmarshal(connectionConfig); err != nil {
		panic(err)
	}

	return connectionConfig
}
