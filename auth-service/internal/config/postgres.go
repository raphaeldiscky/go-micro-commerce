package config

import (
	"log"

	"github.com/spf13/viper"
)

// PostgresConfig holds Postgres configuration.
type PostgresConfig struct {
	Host            string `mapstructure:"POSTGRES_HOST"`
	Port            int    `mapstructure:"POSTGRES_PORT"`
	Name            string `mapstructure:"POSTGRES_DB"`
	User            string `mapstructure:"POSTGRES_USER"`
	Password        string `mapstructure:"POSTGRES_PASSWORD"`
	SSLMode         string `mapstructure:"POSTGRES_SSL_MODE"`
	MaxIdleConns    int    `mapstructure:"POSTGRES_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `mapstructure:"POSTGRES_MAX_OPEN_CONNS"`
	MaxConnLifetime int    `mapstructure:"DB_MAX_CONN_LIFETIME"`
}

// initPostgresConfig initializes the Postgres configuration from environment variables.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", 15432)
	viper.SetDefault("POSTGRES_DB", "auth_db")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_SSL_MODE", "disable")
	viper.SetDefault("POSTGRES_MAX_IDLE_CONNS", 10)
	viper.SetDefault("POSTGRES_MAX_OPEN_CONNS", 32)
	viper.SetDefault("POSTGRES_CONN_MAX_LIFETIME", 60)

	PostgresConfig := &PostgresConfig{}

	if err := viper.Unmarshal(&PostgresConfig); err != nil {
		log.Fatalf("error mapping Postgres config: %v", err)
	}

	return PostgresConfig
}

// GetConnectionString returns the Postgres connection string.
func (d *PostgresConfig) GetConnectionString() string {
	return "postgres://" + d.User + ":" + d.Password + "@" + d.Host + ":" +
		string(rune(d.Port)) + "/" + d.Name + "?sslmode=" + d.SSLMode
}
