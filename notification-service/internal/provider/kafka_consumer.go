package provider

import (
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

// SetupKafkaConsumers initializes the Kafka consumers for the notification service.
func SetupKafkaConsumers(cfg *config.KafkaConfig) []mq.KafkaConsumer {
	// return []mq.KafkaConsumer{
	// 	event.New(
	// 		mq.NewKafkaConsumerGroup(&mq.KafkaConsumerConfig{
	// 			Brokers:       cfg.Brokers,
	// 			ConsumerGroup: constant.OrderCreatedConsumerGroup,
	// 			InitialOffset: sarama.OffsetOldest,
	// 		}),
	// 		dataStore,
	// 	),
	// }
	return []mq.KafkaConsumer{}
}
