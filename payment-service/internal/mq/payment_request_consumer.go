package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
)

// PaymentRequestEvent is the envelope for payment request events.
type PaymentRequestEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.PaymentRequestPayload `json:"payload"`
}

// PaymentRequestConsumer handles payment request events from order service.
type PaymentRequestConsumer struct {
	logger    logger.Logger
	datastore repository.DataStore
}

// NewPaymentRequestConsumer creates a new consumer for payment request events.
func NewPaymentRequestConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
) *PaymentRequestConsumer {
	return &PaymentRequestConsumer{
		logger:    appLogger,
		datastore: ds,
	}
}

// Handler processes payment request events.
func (c *PaymentRequestConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"payment", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		event.PaymentRequestTopic,    // topic
		pkgconstant.OrderServiceName, // source service
		json.RawMessage(body),
		nil, // correlation_id
		nil, // causation_id
	)

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		inboxRepo := ds.InboxRepository()

		// Store event in inbox (handles duplicates automatically)
		storedEvent, err := inboxRepo.Create(ctx, inboxEvent)
		if err != nil {
			return fmt.Errorf("failed to store event in inbox: %w", err)
		}

		// If it's a duplicate, just log and return successfully
		if storedEvent.Status == constant.InboxStatusDuplicate {
			c.logger.Infof(
				"Duplicate payment request event received: %s, skipping processing",
				meta.Metadata.EventID,
			)

			return nil
		}

		// Mark as processing
		if err := inboxRepo.MarkAsProcessing(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processing: %w", err)
		}

		// Process the payment request based on event type
		var processingErr error

		switch meta.Metadata.EventType {
		case event.PaymentRequestedEventType:
			processingErr = c.processPaymentRequest(ctx, ds, body)
		default:
			c.logger.Warnf("ignoring unknown payment event type: %s", meta.Metadata.EventType)
			// Mark as processed even for unknown events to avoid reprocessing
			processingErr = nil
		}

		// Update inbox event status based on processing result
		if processingErr != nil {
			c.logger.Errorf(
				"Failed to process payment request event %s: %v",
				meta.Metadata.EventID,
				processingErr,
			)

			if err := inboxRepo.MarkAsFailed(ctx, storedEvent.ID, processingErr.Error()); err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}

			return processingErr
		}

		if err := inboxRepo.MarkAsProcessed(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processed: %w", err)
		}

		return nil
	})
}

// processPaymentRequest handles payment request events to create payment records.
func (c *PaymentRequestConsumer) processPaymentRequest(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt PaymentRequestEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment request event: %w", err)
	}

	c.logger.Infof("Processing payment request for order ID: %s, amount: %s",
		evt.Payload.OrderID, evt.Payload.TotalPrice)

	paymentRepo := ds.PaymentRepository()
	outboxRepo := ds.OutboxRepository()

	// Check if payment already exists for this order
	existingPayment, err := paymentRepo.FindByOrderID(ctx, evt.Payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		c.logger.Infof(
			"Payment already exists for order %s, skipping creation",
			evt.Payload.OrderID,
		)

		return nil
	}

	paymentMethod, err := mapper.MapStringToPaymentMethod(evt.Payload.PaymentMethod)
	if err != nil {
		return err
	}
	// Create new payment entity
	payment, err := entity.NewPayment(
		evt.Payload.OrderID,
		evt.Payload.TotalPrice,
		evt.Payload.Currency,
		paymentMethod,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment entity: %w", err)
	}

	// Save payment
	savedPayment, err := paymentRepo.Create(ctx, payment)
	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// Create payment created event for the outbox
	paymentEvt := NewPaymentLifecycleEvent(
		savedPayment.OrderID,
		constant.PaymentStatusPending,
		savedPayment.Amount,
	)

	paymentEvtPayload, err := sonic.Marshal(paymentEvt)
	if err != nil {
		return fmt.Errorf("failed to marshal payment event: %w", err)
	}

	outboxEvent := &entity.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "payment",
		AggregateID:   savedPayment.ID,
		EventType:     event.PaymentCreatedEventType,
		Topic:         event.PaymentLifecycleTopic,
		Payload:       paymentEvtPayload,
		Status:        constant.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
		ScheduledFor:  time.Now().UTC(),
		Attempts:      0,
	}

	if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to create payment created event: %w", err)
	}

	c.logger.Infof("Successfully created payment %s for order %s",
		savedPayment.ID, evt.Payload.OrderID)

	return nil
}
