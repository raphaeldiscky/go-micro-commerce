package kafka

import (
	"github.com/IBM/sarama"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// AdminConfig holds the configuration for the Kafka admin.
type AdminConfig struct {
	Brokers []string
}

// Admin represents a Kafka admin client.
type Admin struct {
	client sarama.Client
	logger logger.Logger
}

// NewAdmin creates a new instance of Admin.
func NewAdmin(opt *AdminConfig, logger logger.Logger) (*Admin, error) {
	client, err := sarama.NewClient(opt.Brokers, sarama.NewConfig())
	if err != nil {
		return nil, err
	}

	return &Admin{
		client: client,
		logger: logger,
	}, nil
}

// CreateTopic creates a new Kafka topic.
func (admin *Admin) CreateTopic(topic string, numPartitions, replicationFactor int) error {
	adminClient, err := sarama.NewClusterAdminFromClient(admin.client)
	if err != nil {
		return err
	}

	topics, err := adminClient.ListTopics()
	if err != nil {
		return err
	}

	if _, exists := topics[topic]; !exists {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     int32(numPartitions),
			ReplicationFactor: int16(replicationFactor),
		}
		if errTopic := adminClient.CreateTopic(topic, topicDetail, false); errTopic != nil {
			admin.logger.Fatalf("failed to create topic: %v", errTopic)
		}
	}

	admin.logger.Info("Kafka topic created successfully:", topic)

	return nil
}
