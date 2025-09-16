package service

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// PaymentReminderTaskService handles payment reminder task processing.
type PaymentReminderTaskService interface {
	ProcessPaymentReminder(ctx context.Context, req *dto.PaymentReminderRequest) error
	ProcessOrderCancellation(ctx context.Context, req *dto.CancelOrderRequest) error
}

// PaymentReminderTaskServiceImpl implements PaymentReminderTaskService.
type PaymentReminderTaskServiceImpl struct {
	notificationProducer kafka.ProducerInterface
	dataStore            repository.DataStore
	logger               logger.Logger
}

// NewPaymentReminderTaskService creates a new payment reminder task service.
func NewPaymentReminderTaskService(
	notificationProducer kafka.ProducerInterface,
	dataStore repository.DataStore,
	logger logger.Logger,
) PaymentReminderTaskService {
	return &PaymentReminderTaskServiceImpl{
		notificationProducer: notificationProducer,
		dataStore:            dataStore,
		logger:               logger,
	}
}

// ProcessPaymentReminder processes a payment reminder task.
func (s *PaymentReminderTaskServiceImpl) ProcessPaymentReminder(
	ctx context.Context,
	req *dto.PaymentReminderRequest,
) error {
	s.logger.Infof(
		"Processing payment reminder for order %s (reminder count: %d/%d)",
		req.OrderID,
		req.ReminderCount,
		req.MaxReminders,
	)

	// Create order entity from DTO data - no need to fetch from database
	order := &entity.Order{
		ID:         req.OrderID,
		TotalPrice: req.TotalPrice,
		Currency:   req.Currency,
	}

	// Determine template and subject based on reminder count
	var (
		templateID pkgconstant.TemplateIDType
		subject    string
	)

	switch {
	case req.ReminderCount >= req.MaxReminders:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = "Final Notice - Payment Required for Your Order"
	case req.ReminderCount == 1:
		templateID = pkgconstant.TemplateOrderPaymentRequired
		subject = "Payment Required - Complete Your Order"
	default:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = "Payment Reminder - Complete Your Order"
	}

	// Create notification event using the existing producer function
	notificationEvent := producer.NewNotificationRequestEvent(
		order,
		req.ReservedProducts,
		req.CustomerEmail,
		"Customer Name", // Customer name - could be retrieved from customer service if needed
		nil,             // No tracking number for payment reminders
		templateID,
		subject,
	)

	// Send notification
	if err := s.notificationProducer.Send(ctx, notificationEvent); err != nil {
		s.logger.Errorf(
			"Failed to send payment reminder notification for order %s: %v",
			req.OrderID,
			err,
		)

		return err
	}

	s.logger.Infof(
		"Payment reminder notification sent successfully for order %s (reminder: %d/%d)",
		req.OrderID,
		req.ReminderCount,
		req.MaxReminders,
	)

	return nil
}

// ProcessOrderCancellation processes an order cancellation task.
func (s *PaymentReminderTaskServiceImpl) ProcessOrderCancellation(
	ctx context.Context,
	req *dto.CancelOrderRequest,
) error {
	s.logger.Infof(
		"Processing order cancellation for order %s (reason: %s)",
		req.OrderID,
		req.Reason,
	)

	// For cancellation, we need to fetch the order to get complete order details
	// since CancelOrderTaskPayload only contains basic information
	orderRepo := s.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		s.logger.Errorf("Failed to get order %s: %v", req.OrderID, err)
		return err
	}

	// Create notification event for order cancellation
	notificationEvent := producer.NewNotificationRequestEvent(
		order,
		nil,
		"",  // Email will be retrieved from customer lookup or order data
		"",  // Customer name will be retrieved from customer lookup
		nil, // No tracking number for cancelled orders
		pkgconstant.TemplateOrderCanceled,
		"Order Cancelled - Payment Timeout",
	)

	// Send notification
	if err = s.notificationProducer.Send(ctx, notificationEvent); err != nil {
		s.logger.Errorf(
			"Failed to send order cancellation notification for order %s: %v",
			req.OrderID,
			err,
		)

		return err
	}

	s.logger.Infof(
		"Order cancellation notification sent successfully for order %s",
		req.OrderID,
	)

	return nil
}
