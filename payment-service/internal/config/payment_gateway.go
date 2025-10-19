package config

import "github.com/spf13/viper"

// PaymentGatewayConfig holds configuration for payment gateway providers.
type PaymentGatewayConfig struct {
	Provider            string
	StripeAPIKey        string
	StripeWebhookSecret string
	DefaultCurrency     string
}

// initPaymentGatewayConfig initializes the payment gateway configuration.
func initPaymentGatewayConfig() *PaymentGatewayConfig {
	return &PaymentGatewayConfig{
		Provider:            viper.GetString("PAYMENT_GATEWAY_PROVIDER"),
		StripeAPIKey:        viper.GetString("STRIPE_API_KEY"),
		StripeWebhookSecret: viper.GetString("STRIPE_WEBHOOK_SECRET"),
		DefaultCurrency:     viper.GetString("PAYMENT_GATEWAY_DEFAULT_CURRENCY"),
	}
}
