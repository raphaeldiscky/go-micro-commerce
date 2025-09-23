package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	App             *AppConfig
	HTTPServer      *HTTPServerConfig
	Postgres        *PostgresConfig
	Redis           *RedisConfig
	Consul          *ConsulConfig
	WebSocketServer *WebSocketServerConfig
	Connection      *ConnectionConfig
}

// LoadConfig loads the configuration from environment variables and config files.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	cfg := &Config{
		App:             initAppConfig(),
		HTTPServer:      initHTTPServerConfig(),
		Postgres:        initPostgresConfig(),
		Redis:           initRedisConfig(),
		Consul:          initConsulConfig(),
		WebSocketServer: initWebSocketServerConfig(),
		Connection:      initConnectionConfig(),
	}

	return cfg, nil
}
