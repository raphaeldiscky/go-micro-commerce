package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
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
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_RETRY_MAX", constant.KafkaRetryMax)
	viper.SetDefault("KAFKA_RETRY_INTERVAL", constant.KafkaRetryInterval)
	viper.SetDefault("KAFKA_FLUSH_FREQUENCY", constant.KafkaFlushFrequency)
	viper.SetDefault("KAFKA_RETURN_SUCCESS", true)
	viper.SetDefault("KAFKA_RETURN_ERRORS", true)

	kafkaConfig := &KafkaConfig{}
	if err := viper.Unmarshal(kafkaConfig); err != nil {
		panic(err)
	}

	// Parse comma-separated KAFKA_BROKERS string from environment variable
	brokersStr := viper.GetString("KAFKA_BROKERS")
	if brokersStr != "" {
		kafkaConfig.Brokers = parseCommaSeparated(brokersStr)
	}

	return kafkaConfig
}
