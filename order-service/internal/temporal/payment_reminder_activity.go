package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// PaymentReminderActivities defines the interface for payment reminder activities.
type PaymentReminderActivities interface {
	SendPaymentReminderActivity(ctx context.Context, req dto.PaymentReminderRequest) error
	CheckPaymentStatusActivity(ctx context.Context, orderID uuid.UUID) (bool, error)
	ExpireOrderPaymentActivity(ctx context.Context, req dto.ExpireOrderPaymentRequest) error
}

// paymentReminderActivities implements payment reminder activities.
type paymentReminderActivities struct {
	dataStore              repository.DataStore
	paymentReminderService service.PaymentReminderServiceInterface
}

// NewPaymentReminderActivities creates a new PaymentReminderActivities instance.
func NewPaymentReminderActivities(
	dataStore repository.DataStore,
	paymentReminderService service.PaymentReminderServiceInterface,
) PaymentReminderActivities {
	return &paymentReminderActivities{
		dataStore:              dataStore,
		paymentReminderService: paymentReminderService,
	}
}

// SendPaymentReminderActivity sends a payment reminder notification to the customer.
func (pra *paymentReminderActivities) SendPaymentReminderActivity(
	ctx context.Context,
	req dto.PaymentReminderRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing SendPaymentReminderActivity",
		"orderID",
		req.OrderID,
		"reminderCount",
		req.ReminderCount,
	)

	// Use the payment reminder task service to process the reminder
	err := pra.paymentReminderService.ProcessPaymentReminder(ctx, &req)
	if err != nil {
		logger.Error(
			"Failed to process payment reminder",
			"orderID", req.OrderID,
			"reminderCount", req.ReminderCount,
			"error", err,
		)

		return fmt.Errorf("failed to process payment reminder: %w", err)
	}

	logger.Info(
		"Payment reminder processed successfully",
		"orderID", req.OrderID,
		"reminderCount", req.ReminderCount,
	)

	return nil
}

// CheckPaymentStatusActivity checks if payment has been received for the order.
func (pra *paymentReminderActivities) CheckPaymentStatusActivity(
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

// ExpireOrderPaymentActivity handles order payment expiration using the service layer.
func (pra *paymentReminderActivities) ExpireOrderPaymentActivity(
	ctx context.Context,
	req dto.ExpireOrderPaymentRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ExpireOrderPaymentActivity", "orderID", req.OrderID)

	// Use the payment reminder task service to process the order expiration
	err := pra.paymentReminderService.ProcessOrderExpirePayment(ctx, &req)
	if err != nil {
		logger.Error(
			"Failed to process order payment expiration",
			"orderID", req.OrderID,
			"error", err,
		)

		return fmt.Errorf("failed to process order payment expiration: %w", err)
	}

	logger.Info(
		"Order payment expiration processed successfully",
		"orderID", req.OrderID,
	)

	return nil
}
