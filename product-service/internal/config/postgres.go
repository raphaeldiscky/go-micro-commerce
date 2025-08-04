package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// PostgresConfig holds the configuration for the PostgreSQL database.
type PostgresConfig struct {
	Host            string `mapstructure:"DB_HOST"`
	DBName          string `mapstructure:"DB_NAME"`
	User            string `mapstructure:"DB_USER"`
	Password        string `mapstructure:"DB_PASSWORD"`
	SSLMode         string `mapstructure:"DB_SSL_MODE"`
	Port            int    `mapstructure:"DB_PORT"`
	MaxIdleConn     int    `mapstructure:"DB_MAX_IDLE_CONN"`
	MaxOpenConn     int    `mapstructure:"DB_MAX_OPEN_CONN"`
	MaxConnLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

// Config holds the application configuration.
func initPostgresConfig() *PostgresConfig {
	pgConfig := &PostgresConfig{}

	if err := viper.Unmarshal(&pgConfig); err != nil {
		log.Fatalf("error mapping database config: %v", err)
	}

	return pgConfig
}

// GetURL returns the Postgres connection URL.
func (c *Config) GetURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.DBName,
		c.Postgres.SSLMode,
	)
}
