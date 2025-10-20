package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
)

// WebhookService defines the interface for webhook operations.
type WebhookService interface {
	// HandleStripeWebhook processes Stripe webhook events
	HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error
}

// webhookService implements the WebhookService.
type webhookService struct {
	dataStore           repository.DataStore
	logger              logger.Logger
	stripeWebhookSecret string
}

// NewWebhookService creates a new instance of webhookService.
func NewWebhookService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	stripeWebhookSecret string,
) WebhookService {
	return &webhookService{
		dataStore:           dataStore,
		logger:              appLogger,
		stripeWebhookSecret: stripeWebhookSecret,
	}
}

// HandleStripeWebhook processes Stripe webhook events.
func (s *webhookService) HandleStripeWebhook(
	ctx context.Context,
	payload []byte,
	signature string,
) error {
	// Verify webhook signature for security
	event, err := webhook.ConstructEvent(payload, signature, s.stripeWebhookSecret)
	if err != nil {
		s.logger.Errorf("Failed to verify Stripe webhook signature: %v", err)
		return httperror.NewBadRequestError("invalid webhook signature")
	}

	s.logger.Infof("Processing Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

	// Route event to appropriate handler
	switch constant.StripeWebhookEventType(event.Type) {
	case constant.StripeEventPaymentIntentSucceeded:
		return s.handlePaymentIntentSucceeded(ctx, event)
	case constant.StripeEventPaymentIntentFailed:
		return s.handlePaymentIntentFailed(ctx, event)
	case constant.StripeEventPaymentIntentCanceled:
		return s.handlePaymentIntentCanceled(ctx, event)
	case constant.StripeEventPaymentIntentRequiresAction:
		return s.handlePaymentIntentRequiresAction(ctx, event)
	case constant.StripeEventChargeRefunded:
		return s.handleChargeRefunded(ctx, event)
	default:
		s.logger.Infof("Unhandled Stripe webhook event type: %s", event.Type)
		return nil // Don't error on unhandled events
	}
}

// handlePaymentIntentSucceeded handles payment_intent.succeeded events.
func (s *webhookService) handlePaymentIntentSucceeded(
	ctx context.Context,
	event stripe.Event,
) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		s.logger.Errorf("Failed to unmarshal PaymentIntent: %v", err)
		return httperror.NewInternalServerError("failed to parse webhook data")
	}

	s.logger.Infof(
		"Payment succeeded: %s, amount: %d %s",
		paymentIntent.ID,
		paymentIntent.Amount,
		paymentIntent.Currency,
	)

	return s.updatePaymentStatus(
		ctx,
		paymentIntent.ID,
		constant.PaymentStatusCompleted,
		paymentIntent.Metadata,
	)
}

// handlePaymentIntentFailed handles payment_intent.failed events.
func (s *webhookService) handlePaymentIntentFailed(
	ctx context.Context,
	event stripe.Event,
) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		s.logger.Errorf("Failed to unmarshal PaymentIntent: %v", err)
		return httperror.NewInternalServerError("failed to parse webhook data")
	}

	failureReason := "unknown"
	if paymentIntent.LastPaymentError != nil {
		failureReason = paymentIntent.LastPaymentError.Msg
	}

	s.logger.Infof(
		"Payment failed: %s, reason: %s",
		paymentIntent.ID,
		failureReason,
	)

	return s.updatePaymentStatus(
		ctx,
		paymentIntent.ID,
		constant.PaymentStatusFailed,
		paymentIntent.Metadata,
	)
}

// handlePaymentIntentCanceled handles payment_intent.canceled events.
func (s *webhookService) handlePaymentIntentCanceled(
	ctx context.Context,
	event stripe.Event,
) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		s.logger.Errorf("Failed to unmarshal PaymentIntent: %v", err)
		return httperror.NewInternalServerError("failed to parse webhook data")
	}

	s.logger.Infof("Payment canceled: %s", paymentIntent.ID)

	return s.updatePaymentStatus(
		ctx,
		paymentIntent.ID,
		constant.PaymentStatusFailed,
		paymentIntent.Metadata,
	)
}

// handlePaymentIntentRequiresAction handles payment_intent.requires_action events.
func (s *webhookService) handlePaymentIntentRequiresAction(
	_ context.Context,
	event stripe.Event,
) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		s.logger.Errorf("Failed to unmarshal PaymentIntent: %v", err)
		return httperror.NewInternalServerError("failed to parse webhook data")
	}

	s.logger.Infof(
		"Payment requires action: %s (3DS or other verification)",
		paymentIntent.ID,
	)

	// Payment stays in processing status, waiting for user action
	return nil
}

// handleChargeRefunded handles charge.refunded events.
func (s *webhookService) handleChargeRefunded(
	ctx context.Context,
	event stripe.Event,
) error {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		s.logger.Errorf("Failed to unmarshal Charge: %v", err)
		return httperror.NewInternalServerError("failed to parse webhook data")
	}

	s.logger.Infof(
		"Charge refunded: %s, amount: %d %s",
		charge.ID,
		charge.AmountRefunded,
		charge.Currency,
	)

	// Extract PaymentIntent ID from charge
	paymentIntentID := charge.PaymentIntent.ID

	return s.updatePaymentStatus(
		ctx,
		paymentIntentID,
		constant.PaymentStatusRefunded,
		charge.Metadata,
	)
}

// updatePaymentStatus updates payment status based on webhook event.
//
//nolint:cyclop // ignore complexity
func (s *webhookService) updatePaymentStatus(
	ctx context.Context,
	gatewayID string,
	newStatus constant.PaymentStatus,
	metadata map[string]string,
) error {
	// Extract transaction ID from metadata
	transactionIDStr, ok := metadata["transaction_id"]
	if !ok {
		s.logger.Errorf(
			"transaction_id not found in webhook metadata for gateway ID: %s",
			gatewayID,
		)

		return httperror.NewBadRequestError("transaction_id not found in metadata")
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		s.logger.Errorf("Invalid transaction_id in webhook metadata: %v", err)
		return httperror.NewBadRequestError("invalid transaction_id")
	}

	return s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Get payment by ID
		payment, errFind := paymentRepo.FindByID(ctx, transactionID)
		if errFind != nil {
			s.logger.Errorf("Payment not found: %s, error: %v", transactionID, errFind)
			return httperror.NewNotFoundError("payment not found")
		}

		if payment == nil {
			return httperror.NewPaymentNotFoundError()
		}

		// Check if status transition is valid
		if payment.Status == newStatus {
			s.logger.Infof(
				"Payment %s already in status %s, skipping update",
				transactionID,
				newStatus,
			)

			return nil
		}

		// Update status
		if err = payment.UpdateStatus(newStatus); err != nil {
			s.logger.Errorf("Failed to update payment status: %v", err)
			return httperror.NewInternalServerError("failed to update payment status")
		}

		// Save updated payment
		updatedPayment, errUpdate := paymentRepo.Update(ctx, payment)
		if errUpdate != nil {
			s.logger.Errorf("Failed to save payment update: %v", errUpdate)
			return httperror.NewInternalServerError("failed to save payment")
		}

		// Determine event type based on new status
		var eventType string

		switch newStatus {
		case constant.PaymentStatusCompleted:
			eventType = kafka.PaymentCompletedEventType
		case constant.PaymentStatusFailed:
			eventType = kafka.PaymentFailedEventType
		case constant.PaymentStatusRefunded:
			eventType = kafka.PaymentRefundedEventType
		case constant.PaymentStatusPending, constant.PaymentStatusProcessing:
			eventType = kafka.PaymentProcessingEventType
		case constant.PaymentStatusTimeout:
			eventType = kafka.PaymentTimeoutEventType
		default:
			eventType = kafka.PaymentProcessingEventType
		}

		// Publish payment status change event
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			updatedPayment.Status,
			updatedPayment.Amount,
		)

		payload, errMarshal := json.Marshal(evt)
		if errMarshal != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedPayment.ID,
			EventType:     eventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment status event")
		}

		s.logger.Infof(
			"Payment %s status updated from %s to %s via webhook",
			transactionID,
			payment.Status,
			newStatus,
		)

		return nil
	})
}
