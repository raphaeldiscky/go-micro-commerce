package config

import (
	"log"

	"github.com/spf13/viper"
)

// PostgresConfig holds Postgres configuration.
type PostgresConfig struct {
	Host            string `mapstructure:"DB_HOST"`
	Port            int    `mapstructure:"DB_PORT"`
	Name            string `mapstructure:"DB_NAME"`
	User            string `mapstructure:"DB_USER"`
	Password        string `mapstructure:"DB_PASSWORD"`
	SSLMode         string `mapstructure:"DB_SSL_MODE"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxConnLifetime int    `mapstructure:"DB_MAX_CONN_LIFETIME"`
}

// initPostgresConfig initializes the Postgres configuration from environment variables.
func initPostgresConfig() *PostgresConfig {
	// Set defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 15432)
	viper.SetDefault("DB_NAME", "postgres")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 32)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 60)

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
