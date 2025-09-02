package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/config"
)

// Providers holds all initialized providers.
type Providers struct {
	KafkaAdmin *kafka.Admin
}

// SetupGlobal initializes and returns the providers.
func SetupGlobal(cfg *config.Config) (*Providers, error) {
	// Setup kafka admin
	kafkaAdmin := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		KafkaAdmin: kafkaAdmin,
	}, nil
}
