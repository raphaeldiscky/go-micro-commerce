package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
)

// Providers holds all initialized providers.
type Providers struct {
	KafkaAdmin *mq.KafkaAdmin
}

// SetupGlobal initializes and returns the providers.
func SetupGlobal(cfg *config.Config) (*Providers, error) {
	// Setup kafka admin
	kafkaAdmin := mq.NewKafkaAdmin(&mq.KafkaAdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		KafkaAdmin: kafkaAdmin,
	}, nil
}
