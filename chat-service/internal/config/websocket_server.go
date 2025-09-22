package config

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// WebsocketServerConfig holds the configuration for the Websocket server.
type WebsocketServerConfig struct {
	Host                 string        `mapstructure:"WEBSOCKET_SERVER_HOST"`
	Port                 int           `mapstructure:"WEBSOCKET_SERVER_PORT"`
	GracePeriod          time.Duration `mapstructure:"WEBSOCKET_SERVER_GRACE_PERIOD"`
	RequestTimeoutPeriod time.Duration `mapstructure:"WEBSOCKET_SERVER_REQUEST_TIMEOUT_PERIOD"`
	ReadTimeout          time.Duration `mapstructure:"WEBSOCKET_SERVER_READ_TIMEOUT"`
	WriteTimeout         time.Duration `mapstructure:"WEBSOCKET_SERVER_WRITE_TIMEOUT"`
	IdleTimeout          time.Duration `mapstructure:"WEBSOCKET_SERVER_IDLE_TIMEOUT"`
	ReadHeaderTimeout    time.Duration `mapstructure:"WEBSOCKET_SERVER_READ_HEADER_TIMEOUT"`
	MaxHeaderBytes       int           `mapstructure:"WEBSOCKET_SERVER_MAX_HEADER_BYTES"`
	HSTSMaxAge           int           `mapstructure:"WEBSOCKET_SERVER_HSTS_MAX_AGE"`
	RateLimiter          rate.Limit    `mapstructure:"WEBSOCKET_SERVER_RATE_LIMITER"`
}

// initWebsocketServerConfig initializes the Websocket server configuration from environment variables.
func initWebsocketServerConfig() *WebsocketServerConfig {
	// Set defaults
	viper.SetDefault("WEBSOCKET_SERVER_HOST", "localhost")
	viper.SetDefault("WEBSOCKET_SERVER_PORT", constant.WebsocketServerPort)
	viper.SetDefault("WEBSOCKET_SERVER_GRACE_PERIOD", constant.WebsocketServerGracePeriod)
	viper.SetDefault(
		"WEBSOCKET_SERVER_REQUEST_TIMEOUT_PERIOD",
		constant.WebsocketServerRequestTimeoutPeriod,
	)
	viper.SetDefault("WEBSOCKET_SERVER_READ_TIMEOUT", constant.WebsocketServerReadTimeout)
	viper.SetDefault("WEBSOCKET_SERVER_WRITE_TIMEOUT", constant.WebsocketServerWriteTimeout)
	viper.SetDefault("WEBSOCKET_SERVER_IDLE_TIMEOUT", constant.WebsocketServerIdleTimeout)
	viper.SetDefault(
		"WEBSOCKET_SERVER_READ_HEADER_TIMEOUT",
		constant.WebsocketServerReadHeaderTimeout,
	)
	viper.SetDefault("WEBSOCKET_SERVER_MAX_HEADER_BYTES", constant.WebsocketServerMaxHeaderBytes)
	viper.SetDefault("WEBSOCKET_SERVER_HSTS_MAX_AGE", constant.WebsocketServerHSTSMaxAge)
	viper.SetDefault("WEBSOCKET_SERVER_RATE_LIMITER", constant.WebsocketServerRateLimiter)

	websocketServerConfig := &WebsocketServerConfig{}
	if err := viper.Unmarshal(websocketServerConfig); err != nil {
		panic(err)
	}

	return websocketServerConfig
}
