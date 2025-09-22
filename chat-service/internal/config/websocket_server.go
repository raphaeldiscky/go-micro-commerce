package config

import (
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// WebsocketServerConfig holds the configuration for the Websocket server.
type WebsocketServerConfig struct {
	Host                 string        `mapstructure:"WS_SERVER_HOST"`
	Port                 int           `mapstructure:"WS_SERVER_PORT"`
	GracePeriod          time.Duration `mapstructure:"WS_SERVER_GRACE_PERIOD"`
	RequestTimeoutPeriod time.Duration `mapstructure:"WS_SERVER_REQUEST_TIMEOUT_PERIOD"`
	ReadTimeout          time.Duration `mapstructure:"WS_SERVER_READ_TIMEOUT"`
	WriteTimeout         time.Duration `mapstructure:"WS_SERVER_WRITE_TIMEOUT"`
	IdleTimeout          time.Duration `mapstructure:"WS_SERVER_IDLE_TIMEOUT"`
	ReadHeaderTimeout    time.Duration `mapstructure:"WS_SERVER_READ_HEADER_TIMEOUT"`
	MaxHeaderBytes       int           `mapstructure:"WS_SERVER_MAX_HEADER_BYTES"`
	HSTSMaxAge           int           `mapstructure:"WS_SERVER_HSTS_MAX_AGE"`
	RateLimiter          rate.Limit    `mapstructure:"WS_SERVER_RATE_LIMITER"`
}

// initWebsocketServerConfig initializes the Websocket server configuration from environment variables.
func initWebsocketServerConfig() *WebsocketServerConfig {
	// Set defaults
	viper.SetDefault("WS_SERVER_HOST", "localhost")
	viper.SetDefault("WS_SERVER_PORT", constant.WsServerPort)
	viper.SetDefault("WS_SERVER_GRACE_PERIOD", constant.WsServerGracePeriod)
	viper.SetDefault(
		"WS_SERVER_REQUEST_TIMEOUT_PERIOD",
		constant.WsServerRequestTimeoutPeriod,
	)
	viper.SetDefault("WS_SERVER_READ_TIMEOUT", constant.WsServerReadTimeout)
	viper.SetDefault("WS_SERVER_WRITE_TIMEOUT", constant.WsServerWriteTimeout)
	viper.SetDefault("WS_SERVER_IDLE_TIMEOUT", constant.WsServerIdleTimeout)
	viper.SetDefault(
		"WS_SERVER_READ_HEADER_TIMEOUT",
		constant.WsServerReadHeaderTimeout,
	)
	viper.SetDefault("WS_SERVER_MAX_HEADER_BYTES", constant.WsServerMaxHeaderBytes)
	viper.SetDefault("WS_SERVER_HSTS_MAX_AGE", constant.WsServerHSTSMaxAge)
	viper.SetDefault("WS_SERVER_RATE_LIMITER", constant.WsServerRateLimiter)

	websocketServerConfig := &WebsocketServerConfig{}
	if err := viper.Unmarshal(websocketServerConfig); err != nil {
		panic(err)
	}

	return websocketServerConfig
}
