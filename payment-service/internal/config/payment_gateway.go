package config

import "github.com/spf13/viper"

// PaymentGatewayConfig holds configuration for payment gateway providers.
type PaymentGatewayConfig struct {
	StripeSecretKey             string `mapstructure:"STRIPE_SECRET_KEY"`
	StripeWebhookEndpointSecret string `mapstructure:"STRIPE_WEBHOOK_ENDPOINT_SECRET"`
}

// initPaymentGatewayConfig initializes the payment gateway configuration.
func initPaymentGatewayConfig() *PaymentGatewayConfig {
	viper.SetDefault("STRIPE_SECRET_KEY", "")
	viper.SetDefault("STRIPE_WEBHOOK_ENDPOINT_SECRET", "")

	gatewayConfig := &PaymentGatewayConfig{}
	if err := viper.Unmarshal(gatewayConfig); err != nil {
		panic(err)
	}

	return gatewayConfig
}
