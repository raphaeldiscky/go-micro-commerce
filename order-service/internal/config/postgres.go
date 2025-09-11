package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// PostgresConfig holds the configuration for the PostgreSQL database.
type PostgresConfig struct {
	Host            string `mapstructure:"POSTGRES_HOST"`
	Name            string `mapstructure:"POSTGRES_DB"`
	User            string `mapstructure:"POSTGRES_USER"`
	Password        string `mapstructure:"POSTGRES_PASSWORD"`
	SSLMode         string `mapstructure:"POSTGRES_SSL_MODE"`
	Port            int    `mapstructure:"POSTGRES_PORT"`
	MaxIdleConns    int    `mapstructure:"POSTGRES_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `mapstructure:"POSTGRES_MAX_OPEN_CONNS"`
	MaxConnLifetime int    `mapstructure:"POSTGRES_CONN_MAX_LIFETIME"`
}

// Config holds the application configuration.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 15433)
	viper.SetDefault("POSTGRES_DB", "order_db")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_SSL_MODE", "disable")
	viper.SetDefault("POSTGRES_MAX_IDLE_CONNS", 10)
	viper.SetDefault("POSTGRES_MAX_OPEN_CONNS", 32)
	viper.SetDefault("POSTGRES_CONN_MAX_LIFETIME", 60)

	pgConfig := &PostgresConfig{}

	if err := viper.Unmarshal(&pgConfig); err != nil {
		slog.Error("error mapping postgres config", "err", err)
	}

	return pgConfig
}
