package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

// AdminConfig holds the configuration for the Kafka admin.
type AdminConfig struct {
	Brokers []string
}

// Admin represents a Kafka admin client.
type Admin struct {
	Client sarama.Client
}

// NewAdmin creates a new instance of Admin.
func NewAdmin(opt *AdminConfig) *Admin {
	client, err := sarama.NewClient(opt.Brokers, sarama.NewConfig())
	if err != nil {
		log.Fatalf("failed to create kafka admin: %v", err)
	}

	log.Println("Kafka admin client created successfully")

	return &Admin{
		Client: client,
	}
}

// CreateTopic creates a new Kafka topic.
func (admin *Admin) CreateTopic(topic string, numPartitions, replicationFactor int) {
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

	log.Println("Kafka topic created successfully:", topic)
}
