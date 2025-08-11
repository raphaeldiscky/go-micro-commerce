package config

import (
	"log"

	"github.com/spf13/viper"
)

// SMTPConfig holds the configuration for the SMTP server.
type SMTPConfig struct {
	Host  string `mapstructure:"SMTP_HOST"`
	Email string `mapstructure:"SMTP_EMAIL"`
	Port  int    `mapstructure:"SMTP_PORT"`
}

// initSMTPConfig initializes the SMTP configuration.
func initSMTPConfig() *SMTPConfig {
	// Set defaults
	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_EMAIL", "no-reply@example.com")
	viper.SetDefault("SMTP_PORT", 587)

	smtpConfig := &SMTPConfig{}

	if err := viper.Unmarshal(&smtpConfig); err != nil {
		log.Fatalf("error mapping smtp config: %v", err)
	}

	return smtpConfig
}
