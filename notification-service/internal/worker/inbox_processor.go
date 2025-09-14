// Package worker provides the inbox processor for handling events with exactly-once delivery.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// InboxProcessor is responsible for processing inbox events with exactly-once delivery.
type InboxProcessor struct {
	dataStore                repository.DataStore
	logger                   logger.Logger
	notificationEventService service.NotificationEventService
	config                   config.InboxProcessorConfig
}

// NewInboxProcessor creates a new instance of InboxProcessor.
func NewInboxProcessor(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	notificationEventService service.NotificationEventService,
	cfg config.InboxProcessorConfig,
) *InboxProcessor {
	return &InboxProcessor{
		dataStore:                dataStore,
		logger:                   appLogger,
		notificationEventService: notificationEventService,
		config:                   cfg,
	}
}

// Start begins the inbox processor's processing and cleanup loops.
func (p *InboxProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting inbox processor")

	// Start processing loop
	go p.processLoop(ctx)

	// Start cleanup loop
	go p.cleanupLoop(ctx)

	p.logger.Info("inbox processor started successfully")

	return nil
}

// Shutdown gracefully stops the inbox processor.
func (p *InboxProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("shutting down inbox processor")
	// Context cancellation will stop the loops
	return nil
}

// Name returns the worker name.
func (p *InboxProcessor) Name() string {
	return "inbox-processor"
}

// processLoop periodically processes pending inbox events.
func (p *InboxProcessor) processLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	p.logger.Infof("starting inbox process loop with interval: %v", p.config.PollInterval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("inbox process loop shutting down")

			return
		case <-ticker.C:
			p.processPendingEvents(ctx)
		}
	}
}

// cleanupLoop periodically cleans up processed inbox events.
func (p *InboxProcessor) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	p.logger.Infof("starting inbox cleanup loop with interval: %v", p.config.CleanupInterval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("inbox cleanup loop shutting down")

			return
		case <-ticker.C:
			p.cleanupProcessedEvents(ctx)
		}
	}
}

// processPendingEvents processes events that are pending.
func (p *InboxProcessor) processPendingEvents(ctx context.Context) {
	events, err := p.dataStore.InboxRepository().GetEventsForProcessing(ctx, p.config.BatchSize)
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

// processEvent processes a single inbox event.
func (p *InboxProcessor) processEvent(ctx context.Context, inboxEvent *entity.InboxEvent) error {
	p.logger.Infof(
		"processing inbox event %s of type %s from topic %s",
		inboxEvent.ID,
		inboxEvent.EventType,
		inboxEvent.Topic,
	)

	// Mark as processing
	if err := p.dataStore.InboxRepository().MarkAsProcessing(ctx, inboxEvent.ID); err != nil {
		return fmt.Errorf("failed to mark event as processing: %w", err)
	}

	// Delegate business logic to the service layer
	if err := p.routeEventToService(ctx, inboxEvent); err != nil {
		p.handleProcessingError(ctx, inboxEvent.ID, "failed to process notification", err)

		return err
	}

	// Mark as processed
	if err := p.dataStore.InboxRepository().MarkAsProcessed(ctx, inboxEvent.ID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	p.logger.Infof("successfully processed inbox event %s", inboxEvent.ID)

	return nil
}

// routeEventToService routes events to the appropriate service method.
func (p *InboxProcessor) routeEventToService(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	// Route based on event type to the service layer
	switch inboxEvent.EventType {
	case kafka.NotificationRequestedEventType:
		return p.notificationEventService.ProcessNotificationRequest(ctx, inboxEvent)
	case kafka.EmailVerificationRequestedEventType:
		return p.notificationEventService.ProcessEmailVerificationRequest(ctx, inboxEvent)
	case kafka.UserVerifiedEventType:
		return p.notificationEventService.ProcessEmailUserVerified(ctx, inboxEvent)
	default:
		p.logger.Warnf("ignoring unknown event type: %s", inboxEvent.EventType)

		return nil
	}
}

// handleProcessingError handles errors that occur during event processing.
func (p *InboxProcessor) handleProcessingError(
	ctx context.Context,
	eventID uuid.UUID,
	contextMsg string,
	err error,
) {
	errorMsg := fmt.Sprintf("%s: %v", contextMsg, err)
	p.logger.Errorf("error processing inbox event %s: %s", eventID, errorMsg)

	// Get current event to check attempts
	inboxEvent, getErr := p.dataStore.InboxRepository().GetEventByID(ctx, eventID)
	if getErr != nil {
		if getErr.Error() == constant.InboxEventNotFoundErrorMessage {
			p.logger.Warnf("event %s not found during error handling, possibly cleaned up", eventID)
			return
		}

		p.logger.Errorf("failed to get event for error handling: %v", getErr)

		return
	}

	if inboxEvent.Attempts < p.config.MaxRetryAttempts {
		// Schedule for retry with exponential backoff
		backoffDuration := time.Duration(1<<inboxEvent.Attempts) * p.config.RetryBackoff
		backoffTime := time.Now().Add(backoffDuration)

		if updateErr := p.dataStore.InboxRepository().ScheduleForRetry(ctx, eventID, errorMsg, backoffTime); updateErr != nil {
			p.logger.Errorf("failed to schedule event for retry: %v", updateErr)

			return
		}

		p.logger.Infof(
			"inbox event %s scheduled for retry at %v (attempt %d)",
			eventID,
			backoffTime,
			inboxEvent.Attempts+1,
		)

		return
	}

	// Mark as permanently failed
	if markErr := p.dataStore.InboxRepository().MarkAsFailed(ctx, eventID, errorMsg); markErr != nil {
		p.logger.Errorf("failed to mark event as failed: %v", markErr)
	}

	p.logger.Warnf(
		"inbox event %s marked as permanently failed after %d attempts",
		eventID,
		inboxEvent.Attempts,
	)
}

// cleanupProcessedEvents removes processed events from the database.
func (p *InboxProcessor) cleanupProcessedEvents(ctx context.Context) {
	p.logger.Infof(
		"starting cleanup of processed inbox events older than %v",
		p.config.RetentionPeriod,
	)

	if err := p.dataStore.InboxRepository().CleanupProcessedEvents(ctx, p.config.RetentionPeriod); err != nil {
		p.logger.Errorf("failed to cleanup processed inbox events: %v", err)
	} else {
		p.logger.Info("inbox cleanup completed successfully")
	}
}
