package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultPostgresPort            = 15435
	defaultPostgresMaxIdleConns    = 10
	defaultPostgresMaxOpenConns    = 32
	defaultPostgresConnMaxLifetime = 60 * time.Second
)

// PostgresConfig holds the configuration for the PostgreSQL database.
type PostgresConfig struct {
	Host            string        `mapstructure:"POSTGRES_HOST"`
	Name            string        `mapstructure:"POSTGRES_DB"`
	User            string        `mapstructure:"POSTGRES_USER"`
	Password        string        `mapstructure:"POSTGRES_PASSWORD"`
	SSLMode         string        `mapstructure:"POSTGRES_SSL_MODE"`
	Port            int           `mapstructure:"POSTGRES_PORT"`
	MaxIdleConns    int           `mapstructure:"POSTGRES_MAX_IDLE_CONNS"`
	MaxOpenConns    int           `mapstructure:"POSTGRES_MAX_OPEN_CONNS"`
	MaxConnLifetime time.Duration `mapstructure:"POSTGRES_CONN_MAX_LIFETIME"`
}

// Config holds the application configuration.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", defaultPostgresPort)
	viper.SetDefault("POSTGRES_DB", "fulfillment_db")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_SSL_MODE", "disable")
	viper.SetDefault("POSTGRES_MAX_IDLE_CONNS", defaultPostgresMaxIdleConns)
	viper.SetDefault("POSTGRES_MAX_OPEN_CONNS", defaultPostgresMaxOpenConns)
	viper.SetDefault("POSTGRES_CONN_MAX_LIFETIME", defaultPostgresConnMaxLifetime)

	pgConfig := &PostgresConfig{}

	if err := viper.Unmarshal(pgConfig); err != nil {
		panic(err)
	}

	return pgConfig
}
