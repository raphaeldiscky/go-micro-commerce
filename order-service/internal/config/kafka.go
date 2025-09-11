package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// KafkaConfig holds the configuration for Kafka.
type KafkaConfig struct {
	Brokers        []string `mapstructure:"KAFKA_BROKERS"`
	RetryMax       int      `mapstructure:"KAFKA_RETRY_MAX"`
	FlushFrequency int      `mapstructure:"KAFKA_FLUSH_FREQUENCY"`
	ReturnSuccess  bool     `mapstructure:"KAFKA_RETURN_SUCCESS"`
	ReturnErrors   bool     `mapstructure:"KAFKA_RETURN_ERRORS"`
}

// initKafkaConfig initializes the Kafka configuration from environment variables.
func initKafkaConfig() *KafkaConfig {
	// Set defaults
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_RETRY_MAX", 3)
	viper.SetDefault("KAFKA_FLUSH_FREQUENCY", 1000)
	viper.SetDefault("KAFKA_RETURN_SUCCESS", true)

	kafkaConfig := &KafkaConfig{}

	if err := viper.Unmarshal(&kafkaConfig); err != nil {
		slog.Error("error mapping kafka config", "err", err)
	}

	return kafkaConfig
}
