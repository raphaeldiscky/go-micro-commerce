package service

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// PaymentReminderService handles payment reminder task processing.
type PaymentReminderService interface {
	ProcessPaymentReminder(ctx context.Context, req *dto.PaymentReminderRequest) error
	ProcessOrderExpirePayment(ctx context.Context, req *dto.ExpireOrderPaymentRequest) error
}

// paymentReminderService implements PaymentReminderService.
type paymentReminderService struct {
	notificationProducer   kafka.Producer
	orderLifecycleProducer kafka.Producer
	dataStore              repository.DataStore
	sagaOrchestrator       saga.Orchestrator
	logger                 logger.Logger
}

// NewPaymentReminderService creates a new payment reminder task service.
func NewPaymentReminderService(
	notificationProducer kafka.Producer,
	orderLifecycleProducer kafka.Producer,
	dataStore repository.DataStore,
	sagaOrchestrator saga.Orchestrator,
	logger logger.Logger,
) PaymentReminderService {
	return &paymentReminderService{
		notificationProducer:   notificationProducer,
		orderLifecycleProducer: orderLifecycleProducer,
		dataStore:              dataStore,
		sagaOrchestrator:       sagaOrchestrator,
		logger:                 logger,
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

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		order, err := orderRepo.FindByID(ctx, req.OrderID)
		if err != nil {
			s.logger.Errorf("Failed to fetch order %s for payment reminder: %v", req.OrderID, err)
			return fmt.Errorf("failed to fetch order: %w", err)
		}

		// Determine template and subject based on reminder count
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

		// Send notification
		if err = s.notificationProducer.Send(ctx, notificationEvent); err != nil {
			s.logger.Errorf(
				"Failed to send payment reminder notification for order %s: %v",
				req.OrderID,
				err,
			)

			return err
		}

		return nil
	})
	if err != nil {
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

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		// Get order repository
		orderRepo := ds.OrderRepository()

		// Fetch updated order for notification
		order, err := orderRepo.FindByID(ctx, req.OrderID)
		if err != nil {
			s.logger.Errorf(
				"Failed to fetch updated order %s for notification: %v",
				req.OrderID,
				err,
			)

			return fmt.Errorf("failed to fetch updated order: %w", err)
		}

		// Check if order can be expired (only pending or processing orders)
		if !order.CanBeCancelled() {
			s.logger.Infof(
				"Order %s cannot be expired in current status: %s, skipping",
				req.OrderID,
				order.Status,
			)

			return nil
		}

		// Update status to canceled due to payment timeout
		if err = order.UpdateStatus(constant.OrderStatusPaymentExpired); err != nil {
			return httperror.NewBadRequestError("failed to expire order entity")
		}

		// Save updated order
		updatedOrder, updateErr := orderRepo.Update(ctx, order)
		if updateErr != nil {
			return httperror.NewInternalServerError("failed to expire order")
		}

		// Publish domain event
		evt := producer.NewOrderLifecycleEvent(
			order.ID,
			constant.OrderStatusPaymentExpired,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)
		if err = s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order expired event")
		}

		s.logger.Infof("Successfully expired order %s due to payment timeout", req.OrderID)

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

		// Trigger saga compensation to release reserved products
		if err = s.sagaOrchestrator.TriggerSagaCompensation(ctx, req.OrderID); err != nil {
			s.logger.Errorf(
				"Failed to trigger saga compensation for order %s: %v",
				req.OrderID,
				err,
			)
			// Log error but don't fail the entire operation since order is already expired
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.logger.Infof(
		"Order Expire notification sent successfully for order %s",
		req.OrderID,
	)

	return nil
}
