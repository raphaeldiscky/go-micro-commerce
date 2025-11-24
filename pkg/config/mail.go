package config

import (
	"github.com/spf13/viper"
)

const defaultMailHogPort = 1025

// MailConfig holds the mail provider configuration.
type MailConfig struct {
	Provider       string `mapstructure:"MAIL_PROVIDER"`
	Host           string `mapstructure:"MAIL_HOST"`
	FromEmail      string `mapstructure:"MAIL_FROM_EMAIL"`
	Port           int    `mapstructure:"MAIL_PORT"`
	SendGridAPIKey string `mapstructure:"MAIL_SENDGRID_API_KEY"`
}

// initMailConfig returns the mail configuration.
func initMailConfig() *MailConfig {
	// Set defaults for local development (MailHog)
	viper.SetDefault("MAIL_PROVIDER", "mailhog")
	viper.SetDefault("MAIL_HOST", "localhost")
	viper.SetDefault("MAIL_FROM_EMAIL", "noreply@example.com")
	viper.SetDefault("MAIL_PORT", defaultMailHogPort)
	viper.SetDefault("MAIL_SENDGRID_API_KEY", "")

	mailConfig := &MailConfig{}

	return mailConfig
}
