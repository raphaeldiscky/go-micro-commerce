// Package server provides a Kafka server implementation for consuming messages from Kafka topics.
package server

import (
	"context"
	"time"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

type KafkaConsumerServer struct {
	ctx       context.Context
	cancel    context.CancelFunc
	appLogger logger.Logger
	consumers []mq.KafkaConsumer
}

func NewKafkaConsumerServer(cfg *config.Config) *KafkaConsumerServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaConsumerServer{
		ctx:       ctx,
		cancel:    cancel,
		consumers: provider.SetupKafkaConsumers(cfg.Kafka),
	}
}

func (s *KafkaConsumerServer) Start() error {
	s.appLogger.Info("Running Kafka consumer server...")
	for _, consumer := range s.consumers {
		go consumer.Consume(s.ctx)
	}
	s.appLogger.Info("Kafka server is running...")
	return nil
}

func (s *KafkaConsumerServer) Shutdown() {
	s.appLogger.Info("Attempting to shut down the Kafka server...")
	time.Sleep(10 * time.Second)

	s.cancel()
	for _, consumer := range s.consumers {
		if err := consumer.Close(); err != nil {
			s.appLogger.Errorf("Error closing consumer %v: %v", consumer.Topic(), err)
		}
	}

	s.appLogger.Info("Kafka server shut down gracefully")
}
