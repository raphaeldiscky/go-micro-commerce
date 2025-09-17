// Package worker provides the entry point for starting the background workers.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OutboxPublisher is responsible for publishing outbox events.
type OutboxPublisher struct {
	dataStore                   repository.DataStore
	logger                      logger.Logger
	orderLifecycleProducer      kafka.Producer
	orderDLQProducer            kafka.Producer
	paymentRequestProducer      kafka.Producer
	paymentDLQProducer          kafka.Producer
	fulfillmentRequestProducer  kafka.Producer
	fulfillmentDLQProducer      kafka.Producer
	notificationRequestProducer kafka.Producer
	notificationDLQProducer     kafka.Producer
	config                      config.OutboxPublisherConfig
	eventRegistry               *kafka.EventRegistry
}

// NewOutboxPublisher creates a new instance of OutboxPublisher.
func NewOutboxPublisher(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	orderLifecycleProducer kafka.Producer,
	orderDLQProducer kafka.Producer,
	paymentRequestProducer kafka.Producer,
	paymentDLQProducer kafka.Producer,
	fulfillmentRequestProducer kafka.Producer,
	fulfillmentDLQProducer kafka.Producer,
	notificationRequestProducer kafka.Producer,
	notificationDLQProducer kafka.Producer,
	cfg config.OutboxPublisherConfig,
	eventRegistry *kafka.EventRegistry,
) *OutboxPublisher {
	return &OutboxPublisher{
		dataStore:                   dataStore,
		logger:                      appLogger,
		orderLifecycleProducer:      orderLifecycleProducer,
		orderDLQProducer:            orderDLQProducer,
		paymentRequestProducer:      paymentRequestProducer,
		paymentDLQProducer:          paymentDLQProducer,
		fulfillmentRequestProducer:  fulfillmentRequestProducer,
		fulfillmentDLQProducer:      fulfillmentDLQProducer,
		notificationRequestProducer: notificationRequestProducer,
		notificationDLQProducer:     notificationDLQProducer,
		config:                      cfg,
		eventRegistry:               eventRegistry,
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
		p.handleProcessingError(ctx, outboxEvent.ID, "failed to unmarshal event payload", err)

		return err
	}

	// Route to the appropriate producer based on topic
	var selectedProducer kafka.Producer

	switch outboxEvent.Topic {
	case kafka.OrderLifecycleTopic:
		selectedProducer = p.orderLifecycleProducer
	case kafka.PaymentRequestTopic:
		selectedProducer = p.paymentRequestProducer
	case kafka.FulfillmentRequestTopic:
		selectedProducer = p.fulfillmentRequestProducer
	case kafka.NotificationRequestTopic:
		selectedProducer = p.notificationRequestProducer
	default:
		return fmt.Errorf("unknown topic: %s", outboxEvent.Topic)
	}

	// Publish to Kafka
	if err = selectedProducer.Send(ctx, kafkaEvent); err != nil {
		p.handleProcessingError(ctx, outboxEvent.ID, "failed to publish event to Kafka", err)

		return err
	}

	// Mark as processed
	if err = p.dataStore.OutboxRepository().MarkAsProcessed(ctx, outboxEvent.ID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	p.logger.Infof("successfully published event %s to topic %s", outboxEvent.ID, outboxEvent.Topic)

	return nil
}

// handleProcessingError handles errors that occur during event processing.
func (p *OutboxPublisher) handleProcessingError(
	ctx context.Context,
	eventID uuid.UUID,
	contextMsg string,
	err error,
) {
	errorMsg := fmt.Sprintf("%s: %v", contextMsg, err)
	p.logger.Errorf("error processing event %s: %s", eventID, errorMsg)

	// Get current event to check attempts
	outboxEvent, getErr := p.dataStore.OutboxRepository().GetEventByID(ctx, eventID)
	if getErr != nil {
		if getErr.Error() == constant.OutboxEventNotFoundErrorMessage {
			p.logger.Warnf("event %s not found during error handling, possibly cleaned up", eventID)
			return
		}

		p.logger.Errorf("failed to get event for error handling: %v", getErr)

		return
	}

	if outboxEvent.Attempts < p.config.MaxRetryAttempts {
		// Schedule for retry with exponential backoff
		backoffDuration := time.Duration(1<<outboxEvent.Attempts) * p.config.RetryBackoff

		backoffTime := time.Now().Add(backoffDuration)
		if updateErr := p.dataStore.OutboxRepository().ScheduleForRetry(ctx, eventID, errorMsg, backoffTime); updateErr != nil {
			p.logger.Errorf("failed to schedule event for retry: %v", updateErr)

			return
		}

		p.logger.Infof(
			"event %s scheduled for retry at %v (attempt %d)",
			eventID,
			backoffTime,
			outboxEvent.Attempts+1,
		)

		return
	}

	// Move to DLQ - route to appropriate DLQ based on topic
	var dlqProducer kafka.Producer

	var evt event.BaseEvent

	switch outboxEvent.Topic {
	case kafka.OrderLifecycleTopic:
		dlqProducer = p.orderDLQProducer
		evt = producer.NewOrderDLQEvent(outboxEvent, pkgconstant.DLQReasonMaxRetriesExceeded)
	case kafka.PaymentRequestTopic:
		dlqProducer = p.paymentDLQProducer
		evt = producer.NewPaymentDLQEvent(outboxEvent, pkgconstant.DLQReasonMaxRetriesExceeded)
	case kafka.FulfillmentRequestTopic:
		dlqProducer = p.fulfillmentDLQProducer
		evt = producer.NewFulfillmentDLQEvent(outboxEvent, pkgconstant.DLQReasonMaxRetriesExceeded)
	case kafka.NotificationRequestTopic:
		dlqProducer = p.notificationDLQProducer
		evt = producer.NewNotificationDLQEvent(outboxEvent, pkgconstant.DLQReasonMaxRetriesExceeded)
	default:
		p.logger.Errorf("unknown topic for DLQ: %s, skipping DLQ send", outboxEvent.Topic)
		// Don't send to any DLQ - just log and mark as failed
		dlqProducer = nil
		evt = nil
	}

	if dlqProducer != nil && evt != nil {
		if dlqErr := dlqProducer.Send(ctx, evt); dlqErr != nil {
			p.logger.Errorf("failed to move event to DLQ: %v", dlqErr)
		}
	}
	// Mark as permanently failed in the database
	if markErr := p.dataStore.OutboxRepository().MarkAsFailed(ctx, eventID, errorMsg); markErr != nil {
		p.logger.Errorf("failed to mark event as failed: %v", markErr)
	}

	p.logger.Warnf(
		"event %s marked as permanently failed after %d attempts",
		eventID,
		outboxEvent.Attempts,
	)
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
