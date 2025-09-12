package config

import (
	"github.com/spf13/viper"
)

const (
	defaultSMTPPort = 1025
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
	viper.SetDefault("SMTP_EMAIL", "zundria.putra@gmail.com")
	viper.SetDefault("SMTP_PORT", defaultSMTPPort)

	smtpConfig := &SMTPConfig{}

	if err := viper.Unmarshal(smtpConfig); err != nil {
		panic(err)
	}

	return smtpConfig
}
