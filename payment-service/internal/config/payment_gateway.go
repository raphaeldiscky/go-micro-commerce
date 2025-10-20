package config

import "github.com/spf13/viper"

// PaymentGatewayConfig holds configuration for payment gateway providers.
type PaymentGatewayConfig struct {
	StripeAPIKey        string `mapstructure:"STRIPE_API_KEY"`
	StripeWebhookSecret string `mapstructure:"STRIPE_WEBHOOK_SECRET"`
}

// initPaymentGatewayConfig initializes the payment gateway configuration.
func initPaymentGatewayConfig() *PaymentGatewayConfig {
	viper.SetDefault("STRIPE_API_KEY", "")
	viper.SetDefault("STRIPE_WEBHOOK_SECRET", "")

	gatewayConfig := &PaymentGatewayConfig{}
	if err := viper.Unmarshal(gatewayConfig); err != nil {
		panic(err)
	}

	return gatewayConfig
}
