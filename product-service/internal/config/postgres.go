package config

import (
	"log"

	"github.com/spf13/viper"
)

// PostgresConfig holds the configuration for the PostgreSQL database.
type PostgresConfig struct {
	Host            string `mapstructure:"DB_HOST"`
	Name            string `mapstructure:"DB_NAME"`
	User            string `mapstructure:"DB_USER"`
	Password        string `mapstructure:"DB_PASSWORD"`
	SSLMode         string `mapstructure:"DB_SSL_MODE"`
	Port            int    `mapstructure:"DB_PORT"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxConnLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

// Config holds the application configuration.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 25432)
	viper.SetDefault("DB_NAME", "postgres")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 32)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 60)

	pgConfig := &PostgresConfig{}

	if err := viper.Unmarshal(&pgConfig); err != nil {
		log.Fatalf("error mapping database config: %v", err)
	}

	return pgConfig
}
