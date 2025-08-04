package config

import (
	"log"

	"github.com/spf13/viper"
)

// KafkaConfig holds the configuration for Kafka.
type KafkaConfig struct {
	Brokers        []string `mapstructure:"KAFKA_BROKERS"`
	Topic          string   `mapstructure:"KAFKA_TOPIC"`
	RetryMax       int      `mapstructure:"KAFKA_RETRY_MAX"`
	FlushFrequency int      `mapstructure:"KAFKA_FLUSH_FREQUENCY"`
	ReturnSuccess  bool     `mapstructure:"KAFKA_RETURN_SUCCESS"`
}

// initKafkaConfig initializes the Kafka configuration from environment variables.
func initKafkaConfig() *KafkaConfig {
	kafkaConfig := &KafkaConfig{}

	if err := viper.Unmarshal(&kafkaConfig); err != nil {
		log.Fatalf("error mapping kafka config: %v", err)
	}

	return kafkaConfig
}
