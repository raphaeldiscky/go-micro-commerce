package mq

import (
	"log"

	"github.com/IBM/sarama"
)

// KafkaAdminConfig holds the configuration for the Kafka admin.
type KafkaAdminConfig struct {
	Brokers []string
}

// KafkaAdmin represents a Kafka admin client.
type KafkaAdmin struct {
	Client sarama.Client
}

// NewKafkaAdmin creates a new instance of KafkaAdmin.
func NewKafkaAdmin(opt *KafkaAdminConfig) *KafkaAdmin {
	client, err := sarama.NewClient(opt.Brokers, sarama.NewConfig())
	if err != nil {
		log.Fatalf("failed to create kafka admin: %v", err)
	}

	return &KafkaAdmin{
		Client: client,
	}
}

// CreateTopic creates a new Kafka topic.
func (admin *KafkaAdmin) CreateTopic(topic string, numPartitions, replicationFactor int) {
	adminClient, err := sarama.NewClusterAdminFromClient(admin.Client)
	if err != nil {
		log.Fatalf("failed to create kafka admin: %v", err)
	}

	topics, err := adminClient.ListTopics()
	if err != nil {
		log.Fatalf("failed to list topics: %v", err)
	}

	if _, exists := topics[topic]; !exists {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     int32(numPartitions),
			ReplicationFactor: int16(replicationFactor),
		}
		if err := adminClient.CreateTopic(topic, topicDetail, false); err != nil {
			log.Fatalf("failed to create topic: %v", err)
		}
	}
}
