package config

import (
	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// SMTPConfig holds the SMTP server configuration.
type SMTPConfig struct {
	Host  string `mapstructure:"SMTP_HOST"`
	Email string `mapstructure:"SMTP_EMAIL"`
	Port  int    `mapstructure:"SMTP_PORT"`
}

// initSMTPConfig returns the SMTP configuration.
func initSMTPConfig() *SMTPConfig {
	// Set defaults
	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_EMAIL", "zundria.putra@gmail.com")
	viper.SetDefault("SMTP_PORT", constant.SMTPPort)

	smtpConfig := &SMTPConfig{}

	return smtpConfig
}
