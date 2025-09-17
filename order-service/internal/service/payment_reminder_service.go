package service

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// PaymentReminderService handles payment reminder task processing.
type PaymentReminderService interface {
	ProcessPaymentReminder(ctx context.Context, req *dto.PaymentReminderRequest) error
	ProcessOrderExpirePayment(ctx context.Context, req *dto.ExpireOrderPaymentRequest) error
}

// paymentReminderService implements PaymentReminderService.
type paymentReminderService struct {
	notificationProducer kafka.Producer
	dataStore            repository.DataStore
	orderService         OrderService
	logger               logger.Logger
}

// NewPaymentReminderService creates a new payment reminder task service.
func NewPaymentReminderService(
	notificationProducer kafka.Producer,
	dataStore repository.DataStore,
	orderService OrderService,
	logger logger.Logger,
) PaymentReminderService {
	return &paymentReminderService{
		notificationProducer: notificationProducer,
		dataStore:            dataStore,
		orderService:         orderService,
		logger:               logger,
	}
}

// ProcessPaymentReminder processes a payment reminder task.
func (s *paymentReminderService) ProcessPaymentReminder(
	ctx context.Context,
	req *dto.PaymentReminderRequest,
) error {
	s.logger.Infof(
		"Processing payment reminder for order %s (reminder count: %d/%d)",
		req.OrderID,
		req.ReminderCount,
	)

	s.logger.Infof(
		"====REMINDER 1====, payment reminder request: %+v",
		req,
	)
	// Fetch order from database to get complete order information including items
	orderRepo := s.dataStore.OrderRepository()

	s.logger.Infof(
		"====REMINDER 2====, orderRepo: %+v",
		orderRepo,
	)

	order, err := orderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		s.logger.Errorf("Failed to fetch order %s for payment reminder: %v", req.OrderID, err)
		return fmt.Errorf("failed to fetch order: %w", err)
	}

	s.logger.Infof(
		"====REMINDER 3====, order: %+v",
		order,
	)

	// Determine template and subject based on reminder count
	var (
		templateID pkgconstant.TemplateIDType
		subject    string
	)

	s.logger.Infof(
		"====REMINDER 4====, req.ReminderCount: %d, req.MaxReminders: %d",
		req.ReminderCount,
	)

	switch req.ReminderCount {
	case constant.FirstReminderSequence:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = constant.FirstPaymentReminderEmailSubject
	case constant.SecondReminderSequence:
		templateID = pkgconstant.TemplateOrderPaymentReminder
		subject = constant.SecondPaymentReminderEmailSubject
	default:
		s.logger.Warnf(
			"Invalid reminder count %d for order %s",
			req.ReminderCount,
			req.OrderID,
		)
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

	s.logger.Infof(
		"====REMINDER 4====, notificationEvent: %+v",
		notificationEvent,
	)

	// Send notification
	if err = s.notificationProducer.Send(ctx, notificationEvent); err != nil {
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
	)

	return nil
}

// ProcessOrderExpirePayment processes an order Expire task.
func (s *paymentReminderService) ProcessOrderExpirePayment(
	ctx context.Context,
	req *dto.ExpireOrderPaymentRequest,
) error {
	s.logger.Infof(
		"Processing order Expire for order %s (reason: %s)",
		req.OrderID,
	)

	// Update order status to canceled
	order, err := s.orderService.ExpireOrderPayment(ctx, req)
	if err != nil {
		s.logger.Errorf(
			"Failed to update order status to canceled for order %s: %v",
			req.OrderID,
			err,
		)

		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Create notification event for order Expire
	notificationEvent := producer.NewNotificationRequestEvent(
		order,
		nil,
		req.CustomerEmail, // Use customer email from cancel request
		"Customer Name",   // Customer name - could be retrieved from customer service if needed
		nil,               // No tracking number for expired orders
		pkgconstant.TemplateOrderPaymentExpired,
		constant.OrderPaymentExpiredEmailSubject,
	)

	// Send notification
	if err = s.notificationProducer.Send(ctx, notificationEvent); err != nil {
		s.logger.Errorf(
			"Failed to send order Expire notification for order %s: %v",
			req.OrderID,
			err,
		)

		return err
	}

	s.logger.Infof(
		"Order Expire notification sent successfully for order %s",
		req.OrderID,
	)

	return nil
}
