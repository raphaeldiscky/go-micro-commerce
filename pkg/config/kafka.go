package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	// Default Kafka configuration values.
	defaultKafkaRetryMax       = 3
	defaultKafkaRetryInterval  = 2 * time.Second // retry every 2 seconds
	defaultKafkaFlushFrequency = 1000
)

// KafkaConfig holds the configuration for Kafka.
type KafkaConfig struct {
	Brokers        []string      `mapstructure:"KAFKA_BROKERS"`
	RetryMax       int           `mapstructure:"KAFKA_RETRY_MAX"`
	RetryInterval  time.Duration `mapstructure:"KAFKA_RETRY_INTERVAL"`
	FlushFrequency int           `mapstructure:"KAFKA_FLUSH_FREQUENCY"`
	ReturnSuccess  bool          `mapstructure:"KAFKA_RETURN_SUCCESS"`
	ReturnErrors   bool          `mapstructure:"KAFKA_RETURN_ERRORS"`
}

// initKafkaConfig initializes the Kafka configuration from environment variables.
func initKafkaConfig() *KafkaConfig {
	// Set defaults
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_RETRY_MAX", defaultKafkaRetryMax)
	viper.SetDefault("KAFKA_RETRY_INTERVAL", defaultKafkaRetryInterval)
	viper.SetDefault("KAFKA_FLUSH_FREQUENCY", defaultKafkaFlushFrequency)
	viper.SetDefault("KAFKA_RETURN_SUCCESS", true)

	kafkaConfig := &KafkaConfig{}
	if err := viper.Unmarshal(kafkaConfig); err != nil {
		panic(err)
	}

	return kafkaConfig
}
