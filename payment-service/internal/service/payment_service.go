// Package service provides business logic for payment operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/redisutils"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/repository"
)

// PaymentServiceInterface defines the interface for payment business operations.
type PaymentServiceInterface interface {
	// CreatePayment creates a new payment record from order information
	CreatePayment(ctx context.Context, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	// ProcessPayment processes a payment transaction
	ProcessPayment(
		ctx context.Context,
		paymentID uuid.UUID,
		req dto.ProcessPaymentRequest,
	) (*dto.PaymentResponse, error)
	// GetPaymentByOrderID retrieves payment by order ID
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*dto.PaymentResponse, error)
	// HandleOrderPaymentRequested handles payment requests from order service
	HandleOrderPaymentRequested(
		ctx context.Context,
		orderID uuid.UUID,
		amount decimal.Decimal,
	) error
}

// PaymentService implements the PaymentServiceInterface.
type PaymentService struct {
	dataStore                repository.DataStore
	logger                   logger.Logger
	paymentLifecycleProducer mq.KafkaProducerInterface
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	paymentLifecycleProducer mq.KafkaProducerInterface,
) PaymentServiceInterface {
	return &PaymentService{
		dataStore:                dataStore,
		logger:                   appLogger,
		paymentLifecycleProducer: paymentLifecycleProducer,
	}
}

// CreatePayment creates a new payment record from order information.
func (s *PaymentService) CreatePayment(
	ctx context.Context,
	req dto.CreatePaymentRequest,
) (*dto.PaymentResponse, error) {
	res := new(dto.PaymentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if payment already exists for this order
		existingPayment, err := paymentRepo.FindByOrderID(ctx, req.OrderID)
		if err != nil {
			return httperror.NewInternalServerError("failed to check existing payment")
		}

		if existingPayment != nil {
			// Payment already exists, return existing payment
			res = dto.MapToPaymentResponse(existingPayment)

			return nil
		}

		// Create new payment entity
		payment, err := entity.NewPayment(req.OrderID, req.Amount, req.Currency, req.PaymentMethod)
		if err != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("failed to create payment: %v", err))
		}

		// Save payment
		savedPayment, err := paymentRepo.Create(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to save payment")
		}

		// Publish payment created event
		evt := event.NewPaymentLifecycleEvent(
			savedPayment.ID,
			constant.PaymentStatusPending,
			savedPayment.Amount,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   savedPayment.ID,
			EventType:     constant.KafkaEventTypePaymentCreated,
			Topic:         constant.TopicPaymentLifecycle,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		res = dto.MapToPaymentResponse(savedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

//nolint:gocyclo,revive,cyclop // ignore complexity, ProcessPayment processes a payment transaction is large but intentional.
func (s *PaymentService) ProcessPayment(
	ctx context.Context,
	paymentID uuid.UUID,
	req dto.ProcessPaymentRequest,
) (*dto.PaymentResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreatePaymentTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreatePaymentRetryInterval),
			constant.CreatePaymentRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	res := new(dto.PaymentResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Get payment
		payment, err := paymentRepo.FindByID(ctx, paymentID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get payment")
		}

		if payment == nil {
			return httperror.NewPaymentNotFoundError()
		}

		// Check if payment can be processed
		if !payment.CanBeProcessed() {
			return httperror.NewBadRequestError("payment cannot be processed in current status")
		}

		// Update status to processing
		if err := payment.UpdateStatus(constant.PaymentStatusProcessing); err != nil {
			return httperror.NewBadRequestError("failed to update payment status")
		}

		// Simulate payment gateway processing
		paymentSuccess := s.processWithPaymentGateway(ctx, payment, req.PaymentMethod)

		var finalStatus constant.PaymentStatus
		if paymentSuccess {
			finalStatus = constant.PaymentStatusCompleted
			// Set gateway reference (simulated)
			gateway := "stripe"
			referenceID := fmt.Sprintf("pi_%s", uuid.New().String()[:8])
			gatewayResponse := map[string]interface{}{
				"status":   "succeeded",
				"amount":   payment.Amount.String(),
				"currency": payment.Currency,
			}

			if err := payment.SetGatewayReference(gateway, referenceID, gatewayResponse); err != nil {
				return httperror.NewInternalServerError("failed to set gateway reference")
			}
		} else {
			finalStatus = constant.PaymentStatusFailed

			gatewayResponse := map[string]interface{}{
				"status": "failed",
				"error":  "payment_declined",
			}
			if err := payment.SetGatewayReference("stripe", "", gatewayResponse); err != nil {
				return httperror.NewInternalServerError("failed to set gateway response")
			}
		}

		// Update final status
		if err := payment.UpdateStatus(finalStatus); err != nil {
			return httperror.NewBadRequestError("failed to update final payment status")
		}

		// Save updated payment
		updatedPayment, err := paymentRepo.Update(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// Publish payment completion event
		evt := event.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			finalStatus,
			updatedPayment.Amount,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		eventType := constant.KafkaEventTypePaymentPaid
		if finalStatus == constant.PaymentStatusFailed {
			eventType = "PaymentFailed"
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedPayment.ID,
			EventType:     eventType,
			Topic:         constant.TopicPaymentLifecycle,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment completion event")
		}

		res = dto.MapToPaymentResponse(updatedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetPaymentByOrderID retrieves payment by order ID.
func (s *PaymentService) GetPaymentByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.PaymentResponse, error) {
	paymentRepo := s.dataStore.PaymentRepository()

	payment, err := paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get payment")
	}

	if payment == nil {
		return nil, httperror.NewPaymentNotFoundError()
	}

	return dto.MapToPaymentResponse(payment), nil
}

// HandleOrderPaymentRequested handles payment requests from order service.
func (s *PaymentService) HandleOrderPaymentRequested(
	ctx context.Context,
	orderID uuid.UUID,
	amount decimal.Decimal,
) error {
	// Create payment record for the order
	req := dto.CreatePaymentRequest{
		OrderID:       orderID,
		Amount:        amount,
		Currency:      "IDR",                            // Default currency
		PaymentMethod: constant.PaymentMethodCreditCard, // Default payment method
	}

	_, err := s.CreatePayment(ctx, req)
	if err != nil {
		s.logger.Errorf("Failed to create payment for order %s: %v", orderID, err)

		return err
	}

	s.logger.Infof("Successfully created payment record for order %s", orderID)

	return nil
}

// processWithPaymentGateway simulates payment gateway processing.
func (s *PaymentService) processWithPaymentGateway(
	_ context.Context,
	payment *entity.Payment,
	method constant.PaymentMethod,
) bool {
	// Simulate payment gateway call
	// In real implementation, this would call external payment providers like Stripe, PayPal, etc.
	s.logger.Infof(
		"Processing payment %s with method %s and amount %s",
		payment.ID,
		method,
		payment.Amount,
	)

	// Simulate 90% success rate
	return payment.Amount.LessThan(
		decimal.NewFromFloat(1000000),
	) // Fail payments over 1M IDR for demo
}
