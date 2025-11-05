package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
)

// PaymentRequestEvent is the envelope for payment request events.
type PaymentRequestEvent struct {
	Metadata kafkaevent.Metadata              `json:"metadata"`
	Payload  kafkaevent.PaymentRequestPayload `json:"payload"`
}

// PaymentRequestConsumer handles payment request events from order service.
type PaymentRequestConsumer struct {
	logger         logger.Logger
	datastore      repository.DataStore
	gatewayFactory *gateway.Factory
}

// NewPaymentRequestConsumer creates a new consumer for payment request events.
func NewPaymentRequestConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
	gatewayFactory *gateway.Factory,
) *PaymentRequestConsumer {
	return &PaymentRequestConsumer{
		logger:         appLogger,
		datastore:      ds,
		gatewayFactory: gatewayFactory,
	}
}

// Handler processes payment request events.
func (c *PaymentRequestConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata kafkaevent.Metadata `json:"metadata"`
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
		kafka.PaymentRequestTopic, // topic
		meta.Metadata.Source,
		sonic.NoCopyRawMessage(body),
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
		if err = inboxRepo.MarkAsProcessing(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processing: %w", err)
		}

		// Process the payment request based on event type
		var processingErr error

		switch meta.Metadata.EventType {
		case kafka.PaymentRequestedEventType:
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

			if err = inboxRepo.MarkAsFailed(ctx, storedEvent.ID, processingErr.Error()); err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}

			return processingErr
		}

		if err = inboxRepo.MarkAsProcessed(ctx, storedEvent.ID); err != nil {
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
	if err != nil && err.Error() != constant.PaymentNotFoundErrorMessage {
		return fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		c.logger.Infof(
			"Payment already exists for order %s, skipping creation",
			evt.Payload.OrderID,
		)

		return nil
	}

	paymentGateway, err := mapper.MapStringToPaymentGateway(evt.Payload.PaymentGateway)
	if err != nil {
		return err
	}

	// Create new payment entity
	payment, err := entity.NewPayment(
		evt.Payload.OrderID,
		evt.Payload.TotalPrice,
		evt.Payload.Currency,
		paymentGateway,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment entity: %w", err)
	}

	// Save payment
	savedPayment, err := paymentRepo.Create(ctx, payment)
	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// Create Stripe PaymentIntent immediately to get client_secret for 24h payment window
	clientSecret := c.createPaymentIntent(ctx, paymentRepo, savedPayment, evt.Payload.CustomerID)

	savedPayment, err = paymentRepo.FindByID(ctx, savedPayment.ID)
	if err != nil {
		c.logger.Errorf("Failed to refresh payment after PaymentIntent creation: %v", err)
	}

	// Create payment created event for the outbox
	paymentCreatedEvt := producer.NewPaymentLifecycleEvent(
		savedPayment.ID,
		savedPayment.OrderID,
		constant.PaymentStatusPending,
		savedPayment.Amount,
		clientSecret,           // Stripe client secret for Payment Element
		savedPayment.ExpiresAt, // 24-hour payment window expiry
	)

	paymentCreatedPayload, err := sonic.Marshal(paymentCreatedEvt)
	if err != nil {
		return fmt.Errorf("failed to marshal payment created event: %w", err)
	}

	createdOutboxEvent := &entity.OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "payment",
		AggregateID:   savedPayment.ID,
		EventType:     kafka.PaymentCreatedEventType,
		Topic:         kafka.PaymentLifecycleTopic,
		Payload:       paymentCreatedPayload,
		Status:        constant.OutboxStatusPending,
		CreatedAt:     time.Now().UTC(),
		ScheduledFor:  time.Now().UTC(),
		Attempts:      0,
	}

	if err = outboxRepo.Create(ctx, createdOutboxEvent); err != nil {
		return fmt.Errorf("failed to create payment created event: %w", err)
	}

	c.logger.Infof("Successfully created payment %s for order %s",
		savedPayment.ID, evt.Payload.OrderID)

	return nil
}

// createPaymentIntent creates a payment intent with the gateway and returns the client secret.
func (c *PaymentRequestConsumer) createPaymentIntent(
	ctx context.Context,
	paymentRepo repository.PaymentRepository,
	payment *entity.Payment,
	customerID uuid.UUID,
) *string {
	gatewayClient, err := c.gatewayFactory.GetGateway(string(payment.PaymentGateway))
	if err != nil {
		c.logger.Warnf(
			"Failed to get payment gateway client for %s: %v",
			payment.PaymentGateway,
			err,
		)

		return nil
	}

	c.logger.Infof(
		"Creating PaymentIntent for payment %s with gateway %s",
		payment.ID,
		payment.PaymentGateway,
	)

	gatewayReq := &dto.PaymentGatewayRequest{
		PaymentID:  payment.ID,
		CustomerID: customerID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		ExpiresAt:  payment.ExpiresAt,
	}

	gwResp, gwErr := gatewayClient.ProcessPayment(ctx, gatewayReq)
	if gwErr != nil {
		c.logger.Errorf(
			"Failed to create PaymentIntent for payment %s: %v",
			payment.ID,
			gwErr,
		)

		return nil
	}

	c.logger.Infof(
		"Successfully created PaymentIntent for payment %s with client_secret",
		payment.ID,
	)

	if err = payment.SetGatewayReference(
		payment.PaymentGateway,
		gwResp.PaymentIntentID,
		gwResp.GatewayResponse,
	); err != nil {
		c.logger.Errorf("Failed to set gateway reference: %v", err)
		return gwResp.ClientSecret
	}

	if _, err = paymentRepo.Update(ctx, payment); err != nil {
		c.logger.Errorf("Failed to update payment with gateway reference: %v", err)
	}

	return gwResp.ClientSecret
}
