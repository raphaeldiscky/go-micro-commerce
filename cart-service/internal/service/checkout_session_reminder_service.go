// Package service provides business logic for checkout session reminder operations.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
)

// CheckoutSessionReminderService defines the interface for checkout session reminder operations.
type CheckoutSessionReminderService interface {
	ProcessCheckoutSessionReminder(
		ctx context.Context,
		req *dto.CheckoutSessionReminderRequest,
	) error
}

// checkoutSessionReminderService implements the CheckoutSessionReminderService.
type checkoutSessionReminderService struct {
	dataStore repository.DataStore
	logger    logger.Logger
}

// NewCheckoutSessionReminderService creates a new instance of checkoutSessionReminderService.
func NewCheckoutSessionReminderService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) CheckoutSessionReminderService {
	return &checkoutSessionReminderService{
		dataStore: dataStore,
		logger:    appLogger,
	}
}

// ProcessCheckoutSessionReminder processes the checkout session reminder task.
func (s *checkoutSessionReminderService) ProcessCheckoutSessionReminder(
	ctx context.Context,
	req *dto.CheckoutSessionReminderRequest,
) error {
	s.logger.Infof(
		"Processing checkout session reminder for session: %s",
		req.CheckoutSessionID,
	)

	checkoutSessionRepo := s.dataStore.CheckoutSessionRepository()

	// Get the checkout session
	session, err := checkoutSessionRepo.GetByID(ctx, req.CheckoutSessionID)
	if err != nil {
		s.logger.Errorf(
			"Failed to get checkout session %s: %v",
			req.CheckoutSessionID,
			err,
		)

		return fmt.Errorf("failed to get checkout session: %w", err)
	}

	if session == nil {
		s.logger.Warnf("Checkout session %s not found, skipping reminder", req.CheckoutSessionID)
		return nil
	}

	// Check if session is still pending
	if session.Status != constant.CheckoutSessionStatusPending {
		s.logger.Infof(
			"Checkout session %s is no longer pending (status: %s), skipping reminder",
			req.CheckoutSessionID,
			session.Status,
		)

		return nil
	}

	// Send push notification via Kafka
	if err = s.sendReminderNotification(ctx, session, req.CustomerEmail); err != nil {
		s.logger.Errorf(
			"Failed to send reminder notification for checkout session %s: %v",
			req.CheckoutSessionID,
			err,
		)

		return fmt.Errorf("failed to send reminder notification: %w", err)
	}

	s.logger.Infof(
		"Successfully processed checkout session reminder for session: %s",
		req.CheckoutSessionID,
	)

	return nil
}

// sendReminderNotification sends a push notification via Kafka to notification-service.
func (s *checkoutSessionReminderService) sendReminderNotification(
	ctx context.Context,
	session *entity.CheckoutSession,
	customerEmail string,
) error {
	s.logger.Infof(
		"Sending checkout session reminder notification for session: %s",
		session.ID,
	)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create notification event
		notificationEvent := producer.NewNotificationRequestEvent(
			session.ID,
			customerEmail,
		)

		payload, err := sonic.Marshal(notificationEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal notification event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "checkout_session",
			AggregateID:   session.ID,
			EventType:     kafka.NotificationRequestedEventType,
			Topic:         kafka.NotificationRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			s.logger.Errorf(
				"Failed to create reminder notification for checkout session %s: %v",
				session.ID,
				err,
			)

			return fmt.Errorf("failed to create reminder notification event: %w", err)
		}

		s.logger.Infof(
			"Successfully created reminder notification for checkout session: %s",
			session.ID,
		)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to send reminder notification: %w", err)
	}

	return nil
}
