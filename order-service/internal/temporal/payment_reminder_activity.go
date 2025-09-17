package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"go.temporal.io/sdk/activity"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// PaymentReminderActivities defines the interface for payment reminder activities.
type PaymentReminderActivities interface {
	SendPaymentReminderActivity(ctx context.Context, req dto.PaymentReminderRequest) error
	CheckPaymentStatusActivity(ctx context.Context, orderID uuid.UUID) (bool, error)
}

// PaymentReminderActivitiesImpl implements payment reminder activities.
type PaymentReminderActivitiesImpl struct {
	dataStore repository.DataStore
}

// NewPaymentReminderActivities creates a new PaymentReminderActivities instance.
func NewPaymentReminderActivities(dataStore repository.DataStore) PaymentReminderActivities {
	return &PaymentReminderActivitiesImpl{
		dataStore: dataStore,
	}
}

// SendPaymentReminderActivity sends a payment reminder notification to the customer.
func (pra *PaymentReminderActivitiesImpl) SendPaymentReminderActivity(
	ctx context.Context,
	req dto.PaymentReminderRequest,
) error {
	logger := activity.GetLogger(ctx)
	orderID := req.OrderID
	logger.Info("Executing SendPaymentReminderActivity", "orderID", orderID)

	orderRepo := pra.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	var (
		templateID pkgconstant.TemplateIDType
		subject    string
	)

	switch req.ReminderCount {
	case constant.FirstReminderSequence:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = constant.FirstPaymentReminderEmailSubject
	case constant.SecondReminderSequence:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = constant.SecondPaymentReminderEmailSubject
	default:
		// do nothing
	}

	err = pra.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create notification event for reminder
		notificationEvent := producer.NewNotificationRequestEvent(
			order,
			nil,
			req.CustomerEmail,
			"Customer", // TODO: Get actual customer name
			nil,        // No tracking number for reminder
			templateID,
			subject,
		)

		payload, marshalErr := json.Marshal(notificationEvent)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal notification event: %w", marshalErr)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "notification",
			AggregateID:   orderID,
			EventType:     kafka.NotificationRequestedEventType,
			Topic:         kafka.NotificationRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			ScheduledFor:  time.Now().UTC(),
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create payment reminder notification event: %w", err)
		}

		logger.Info(
			"Successfully created payment reminder notification",
			"orderID", req.OrderID,
			"reminderCount", req.ReminderCount,
			"template", templateID,
		)

		return nil
	})

	return err
}

// CheckPaymentStatusActivity checks if payment has been received for the order.
func (pra *PaymentReminderActivitiesImpl) CheckPaymentStatusActivity(
	ctx context.Context,
	orderID uuid.UUID,
) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CheckPaymentStatusActivity", "orderID", orderID)

	orderRepo := pra.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return false, fmt.Errorf("failed to get order: %w", err)
	}

	paymentReceived := order.IsPaymentConfirmed()

	logger.Info(
		"Payment status check completed",
		"orderID", orderID,
		"status", order.Status,
		"paymentReceived", paymentReceived,
	)

	return paymentReceived, nil
}
