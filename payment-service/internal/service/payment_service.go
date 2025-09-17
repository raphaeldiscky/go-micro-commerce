// Package service provides business logic for payment operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/utils/redisutils"
)

// PaymentService defines the interface for payment business operations.
type PaymentService interface {
	// CreatePayment creates a new payment record from order information
	CreatePayment(
		ctx context.Context,
		req dto.CreatePaymentRequest,
	) (*dto.PaymentResponse, error)
	// ProcessPayment processes a payment transaction
	ProcessPayment(
		ctx context.Context,
		orderID uuid.UUID,
		req dto.ProcessPaymentRequest,
	) (*dto.PaymentResponse, error)
	// TimeoutPayment times out a payment transaction
	TimeoutPayment(ctx context.Context, orderID uuid.UUID) error
	// GetPaymentByOrderID retrieves payment by order ID
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*dto.PaymentResponse, error)
}

// paymentService implements the PaymentService.
type paymentService struct {
	dataStore                repository.DataStore
	logger                   logger.Logger
	paymentLifecycleProducer kafka.Producer
	bankingClient            client.BankingClient
	paymentGatewayClient     client.PaymentGatewayClient
}

// NewPaymentService creates a new instance of paymentService.
func NewPaymentService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	paymentLifecycleProducer kafka.Producer,
	bankingClient client.BankingClient,
	paymentGatewayClient client.PaymentGatewayClient,
) PaymentService {
	return &paymentService{
		dataStore:                dataStore,
		logger:                   appLogger,
		paymentLifecycleProducer: paymentLifecycleProducer,
		bankingClient:            bankingClient,
		paymentGatewayClient:     paymentGatewayClient,
	}
}

// CreatePayment creates a new payment record from order information.
func (s *paymentService) CreatePayment(
	ctx context.Context,
	req dto.CreatePaymentRequest,
) (*dto.PaymentResponse, error) {
	res := new(dto.PaymentResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if payment already exists for this order
		existingPayment, err := paymentRepo.FindByOrderID(ctx, req.OrderID)
		if err != nil && err.Error() != constant.PaymentNotFoundErrorMessage {
			return httperror.NewInternalServerError("failed to check existing payment")
		}

		if existingPayment != nil {
			// Payment already exists, return existing payment
			res = mapper.MapToPaymentResponse(existingPayment)

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
		evt := producer.NewPaymentLifecycleEvent(
			savedPayment.ID,
			savedPayment.OrderID,
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
			EventType:     kafka.PaymentCreatedEventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create outbox event")
		}

		res = mapper.MapToPaymentResponse(savedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

//nolint:gocyclo,revive,cyclop,nolintlint // ignore complexity, ProcessPayment processes a payment transaction is large but intentional.
func (s *paymentService) ProcessPayment(
	ctx context.Context,
	orderID uuid.UUID,
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
		if err = lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	res := new(dto.PaymentResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		// Get payment
		payment, errFind := paymentRepo.FindByOrderID(ctx, orderID)
		if errFind != nil {
			if errFind.Error() == constant.PaymentNotFoundErrorMessage {
				return httperror.NewNotFoundError("payment not found for order")
			}

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
		if err = payment.UpdateStatus(constant.PaymentStatusProcessing); err != nil {
			return httperror.NewBadRequestError("failed to update payment status")
		}

		// Process payment with payment gateway
		paymentResult, errProcess := s.processWithPaymentGateway(ctx, payment, req)
		if errProcess != nil {
			return httperror.NewInternalServerError("failed to process payment with gateway")
		}

		var finalStatus constant.PaymentStatus
		if paymentResult.Status == constant.PaymentGatewayStatusSucceeded {
			finalStatus = constant.PaymentStatusCompleted
		} else {
			finalStatus = constant.PaymentStatusFailed
		}

		// Set gateway reference
		if err = payment.SetGatewayReference("stripe", paymentResult.GatewayID, paymentResult.GatewayResponse); err != nil {
			return httperror.NewInternalServerError("failed to set gateway reference")
		}

		// Update final status
		if err = payment.UpdateStatus(finalStatus); err != nil {
			return httperror.NewBadRequestError("failed to update final payment status")
		}

		// Save updated payment
		updatedPayment, errUpdate := paymentRepo.Update(ctx, payment)
		if errUpdate != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// Publish payment completion event
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			finalStatus,
			updatedPayment.Amount,
		)

		payload, errMarshal := json.Marshal(evt)
		if errMarshal != nil {
			return httperror.NewInternalServerError("failed to marshal payment event")
		}

		eventType := kafka.PaymentCompletedEventType
		if finalStatus == constant.PaymentStatusFailed {
			eventType = kafka.PaymentFailedEventType
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
			return httperror.NewInternalServerError("failed to create payment completion event")
		}

		res = mapper.MapToPaymentResponse(updatedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetPaymentByOrderID retrieves payment by order ID.
func (s *paymentService) GetPaymentByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.PaymentResponse, error) {
	paymentRepo := s.dataStore.PaymentRepository()

	payment, err := paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		if err.Error() == constant.PaymentNotFoundErrorMessage {
			return nil, httperror.NewNotFoundError("payment not found for order")
		}

		return nil, httperror.NewInternalServerError("failed to get payment")
	}

	if payment == nil {
		return nil, httperror.NewPaymentNotFoundError()
	}

	return mapper.MapToPaymentResponse(payment), nil
}

// processWithPaymentGateway processes payment with payment gateway client.
func (s *paymentService) processWithPaymentGateway(
	ctx context.Context,
	payment *entity.Payment,
	req dto.ProcessPaymentRequest,
) (*dto.PaymentGatewayResponse, error) {
	s.logger.Infof(
		"Processing payment %s with method %s and amount %s",
		payment.ID,
		req.PaymentMethod,
		payment.Amount,
	)

	paymentRequest := &dto.PaymentGatewayRequest{
		TransactionID:  payment.ID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		PaymentMethod:  req.PaymentMethod,
		CustomerID:     req.CustomerID,
		CustomerEmail:  req.CustomerEmail,
		IdempotencyKey: req.IdempotencyKey.String(),
	}

	return s.paymentGatewayClient.ProcessPayment(ctx, paymentRequest)
}

// TimeoutPayment times out a payment transaction.
func (s *paymentService) TimeoutPayment(ctx context.Context, orderID uuid.UUID) error {
	s.logger.Infof("Processing payment timeout for order: %s", orderID)

	return s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		paymentRepo := ds.PaymentRepository()
		outboxRepo := ds.OutboxRepository()

		payment, err := paymentRepo.FindByOrderID(ctx, orderID)
		if err != nil {
			if err.Error() == constant.PaymentNotFoundErrorMessage {
				s.logger.Warnf("Payment not found for order %s during timeout", orderID)
				return nil // Don't error if payment doesn't exist
			}

			return httperror.NewInternalServerError("failed to get payment")
		}

		if payment == nil {
			s.logger.Warnf("Payment not found for order %s during timeout", orderID)
			return nil // Don't error if payment doesn't exist
		}

		// Check if payment can be timed out (only pending payments)
		if payment.Status != constant.PaymentStatusPending {
			s.logger.Infof(
				"Payment for order %s is already in status %s, skipping timeout",
				orderID,
				payment.Status,
			)

			return nil
		}

		// Update payment status to timed out
		if err = payment.UpdateStatus(constant.PaymentStatusTimeout); err != nil {
			return httperror.NewInternalServerError("failed to update payment status to timeout")
		}

		// Save updated payment
		updatedPayment, err := paymentRepo.Update(ctx, payment)
		if err != nil {
			return httperror.NewInternalServerError("failed to update payment")
		}

		// Publish payment timeout event using outbox pattern
		evt := producer.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			updatedPayment.OrderID,
			constant.PaymentStatusTimeout,
			updatedPayment.Amount,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment timeout event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedPayment.ID,
			EventType:     kafka.PaymentTimeoutEventType,
			Topic:         kafka.PaymentLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment timeout event")
		}

		s.logger.Infof("Successfully timed out payment for order %s", orderID)

		return nil
	})
}
