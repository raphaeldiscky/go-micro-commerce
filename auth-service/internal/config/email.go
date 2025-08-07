package config

import (
	"log"

	"github.com/spf13/viper"
)

// EventPublisherConfig holds event publisher configuration for messaging.
type EventPublisherConfig struct {
	Type           string `mapstructure:"EVENT_PUBLISHER_TYPE"`
	BrokerURL      string `mapstructure:"EVENT_PUBLISHER_BROKER_URL"`
	Topic          string `mapstructure:"EVENT_PUBLISHER_TOPIC"`
	RetryAttempts  int    `mapstructure:"EVENT_PUBLISHER_RETRY_ATTEMPTS"`
	TimeoutSeconds int    `mapstructure:"EVENT_PUBLISHER_TIMEOUT_SECONDS"`
}

// initEventPublisherConfig initializes the event publisher configuration from environment variables.
func initEventPublisherConfig() *EventPublisherConfig {
	// Set defaults
	viper.SetDefault("EVENT_PUBLISHER_TYPE", "kafka")
	viper.SetDefault("EVENT_PUBLISHER_BROKER_URL", "localhost:9092")
	viper.SetDefault("EVENT_PUBLISHER_TOPIC", "auth-events")
	viper.SetDefault("EVENT_PUBLISHER_RETRY_ATTEMPTS", 3)
	viper.SetDefault("EVENT_PUBLISHER_TIMEOUT_SECONDS", 30)

	eventConfig := &EventPublisherConfig{}

	if err := viper.Unmarshal(&eventConfig); err != nil {
		log.Fatalf("error mapping event publisher config: %v", err)
	}

	return eventConfig
}
