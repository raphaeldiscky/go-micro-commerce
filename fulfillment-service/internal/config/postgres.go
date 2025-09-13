package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// PostgresConfig holds the configuration for the PostgreSQL database.
type PostgresConfig struct {
	Host            string        `mapstructure:"POSTGRES_HOST"`
	Port            int           `mapstructure:"POSTGRES_PORT"`
	DB              string        `mapstructure:"POSTGRES_DB"`
	User            string        `mapstructure:"POSTGRES_USER"`
	Password        string        `mapstructure:"POSTGRES_PASSWORD"`
	SSLMode         string        `mapstructure:"POSTGRES_SSL_MODE"`
	MaxIdleConns    int           `mapstructure:"POSTGRES_MAX_IDLE_CONNS"`
	MaxOpenConns    int           `mapstructure:"POSTGRES_MAX_OPEN_CONNS"`
	MaxConnLifetime time.Duration `mapstructure:"POSTGRES_CONN_MAX_LIFETIME"`
}

// Config holds the application configuration.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", constant.PostgresPort)
	viper.SetDefault("POSTGRES_DB", "fulfillment_db")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_SSL_MODE", "disable")
	viper.SetDefault("POSTGRES_MAX_IDLE_CONNS", constant.PostgresMaxIdleConns)
	viper.SetDefault("POSTGRES_MAX_OPEN_CONNS", constant.PostgresMaxOpenConns)
	viper.SetDefault("POSTGRES_MAX_CONN_LIFETIME", constant.PostgresConnMaxLifetime)

	pgConfig := &PostgresConfig{}

	if err := viper.Unmarshal(pgConfig); err != nil {
		panic(err)
	}

	return pgConfig
}
