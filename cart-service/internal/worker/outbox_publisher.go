// Package worker provides the entry point for starting the background workers.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
)

// OutboxPublisher is responsible for publishing outbox events.
type OutboxPublisher struct {
	dataStore                         repository.DataStore
	logger                            logger.Logger
	checkoutSesssionLifecycleProducer kafka.Producer
	config                            config.OutboxPublisherConfig
	eventRegistry                     *kafka.EventRegistry
}

// NewOutboxPublisher creates a new instance of OutboxPublisher.
func NewOutboxPublisher(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	checkoutSesssionLifecycleProducer kafka.Producer,
	cfg config.OutboxPublisherConfig,
	eventRegistry *kafka.EventRegistry,
) *OutboxPublisher {
	return &OutboxPublisher{
		dataStore:                         dataStore,
		logger:                            appLogger,
		checkoutSesssionLifecycleProducer: checkoutSesssionLifecycleProducer,
		config:                            cfg,
		eventRegistry:                     eventRegistry,
	}
}

// Start begins the outbox publisher's processing and cleanup loops.
func (p *OutboxPublisher) Start(ctx context.Context) {
	p.logger.Info("starting outbox publisher")
	// Start processing loop
	go p.processLoop(ctx)

	// Start cleanup loop
	go p.cleanupLoop(ctx)

	p.logger.Info("outbox publisher started successfully")
}

// processLoop periodically processes pending outbox events.
func (p *OutboxPublisher) processLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	p.logger.Infof("starting process loop with interval: %v", p.config.PollInterval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("process loop shutting down")

			return
		case <-ticker.C:
			p.processPendingEvents(ctx)
		}
	}
}

// cleanupLoop periodically cleans up processed outbox events.
func (p *OutboxPublisher) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	p.logger.Infof("starting outbox cleanup loop with interval: %v", p.config.CleanupInterval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("outbox cleanup loop shutting down")

			return
		case <-ticker.C:
			p.cleanupProcessedEvents(ctx)
		}
	}
}

// processPendingEvents processes events that are pending.
func (p *OutboxPublisher) processPendingEvents(ctx context.Context) {
	events, err := p.dataStore.OutboxRepository().GetEventsForProcessing(ctx, p.config.BatchSize)
	if err != nil {
		p.logger.Errorf("failed to get events for processing: %v", err)

		return
	}

	for _, event := range events {
		if err = p.processEvent(ctx, event); err != nil {
			p.logger.Errorf("failed to process event %s: %v", event.ID, err)
		}
	}
}

// processEvent processes a single outbox event.
func (p *OutboxPublisher) processEvent(ctx context.Context, outboxEvent *entity.OutboxEvent) error {
	p.logger.Infof(
		"processing event %s of type %s on topic %s",
		outboxEvent.ID,
		outboxEvent.EventType,
		outboxEvent.Topic,
	)

	// Mark as processing
	if err := p.dataStore.OutboxRepository().MarkAsProcessing(ctx, outboxEvent.ID); err != nil {
		return fmt.Errorf("failed to mark event as processing: %w", err)
	}

	// Parse the event payload
	kafkaEvent, err := p.eventRegistry.UnmarshalEvent(outboxEvent.EventType, outboxEvent.Payload)
	if err != nil {
		return err
	}

	// Route to the appropriate producer based on topic
	var selectedProducer kafka.Producer

	switch outboxEvent.EventType {
	case kafka.CheckoutSessionOrderPlacedEventType:
		selectedProducer = p.checkoutSesssionLifecycleProducer
	default:
		return fmt.Errorf("unknown topic: %s", outboxEvent.Topic)
	}

	// Publish to Kafka
	if err = selectedProducer.Send(ctx, kafkaEvent); err != nil {
		return err
	}

	// Mark as processed
	if err = p.dataStore.OutboxRepository().MarkAsProcessed(ctx, outboxEvent.ID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	p.logger.Infof("successfully published event %s to topic %s", outboxEvent.ID, outboxEvent.Topic)

	return nil
}

// cleanupProcessedEvents removes processed events from the database.
func (p *OutboxPublisher) cleanupProcessedEvents(ctx context.Context) {
	p.logger.Infof("starting cleanup of processed events older than %v", p.config.RetentionPeriod)

	if err := p.dataStore.OutboxRepository().CleanupProcessedEvents(ctx, p.config.RetentionPeriod); err != nil {
		p.logger.Errorf("failed to cleanup processed events: %v", err)
	} else {
		p.logger.Info("cleanup completed successfully")
	}
}
