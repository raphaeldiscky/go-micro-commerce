package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
)

// SetupKafkaConsumers initializes the Kafka consumers for the order service.
func SetupKafkaConsumers(
	_ *config.KafkaConfig,
	_ logger.Logger,
	_ *Providers,
) []kafka.Consumer {
	var consumers []kafka.Consumer

	return consumers
}
