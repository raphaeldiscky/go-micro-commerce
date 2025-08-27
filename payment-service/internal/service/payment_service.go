// Package service provides business logic for order operations.
package service

import (
	"context"
	"fmt"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/utils/redisutils"
)

// PaymentServiceInterface defines the interface for order business operations.
type PaymentServiceInterface interface {
	PayPayment(
		ctx context.Context,
		req dto.PaymentRequest,
		id uuid.UUID,
	) (*dto.PaymentResponse, error)
}

// PaymentService implements the PaymentServiceInterface.
type PaymentService struct {
	dataStore              repository.DataStore
	logger                 logger.Logger
	orderLifecycleProducer mq.KafkaProducerInterface
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	orderLifecycleProducer mq.KafkaProducerInterface,
) PaymentServiceInterface {
	return &PaymentService{
		dataStore:              dataStore,
		logger:                 appLogger,
		orderLifecycleProducer: orderLifecycleProducer,
	}
}

// PayPayment processes payment for an order.
func (s *PaymentService) PayPayment(
	ctx context.Context,
	req dto.PaymentRequest,
	id uuid.UUID,
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

		// Check if order exists
		existingPayment, err := paymentRepo.FindByID(ctx, id)
		if err != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingPayment == nil {
			return httperror.NewPaymentNotFoundError()
		}

		// Check if order can be paid
		if !existingPayment.CanBePaid() {
			return httperror.NewBadRequestError("order cannot be paid in current status")
		}

		if existingPayment.IdempotencyKey == req.IdempotencyKey {
			return httperror.NewBadRequestError(
				fmt.Sprintf(
					"idempotency key for update conflict, existing key: %s , updated key: %s",
					existingPayment.IdempotencyKey,
					req.IdempotencyKey,
				),
			)
		}

		updatedPayment := existingPayment
		updatedPayment.IdempotencyKey = req.IdempotencyKey
		// Update status to paid
		if err := updatedPayment.UpdateStatus(constant.PaymentStatusPaid); err != nil {
			return httperror.NewBadRequestError("failed to update order status entity")
		}

		// Set idempotency key
		updatedPayment.IdempotencyKey = req.IdempotencyKey

		// Save updated order
		updatedPayment, err = paymentRepo.Update(ctx, updatedPayment)
		if err != nil {
			return httperror.NewInternalServerError("failed to update order")
		}

		// Publish domain event
		evt := event.NewPaymentLifecycleEvent(
			updatedPayment.ID,
			constant.PaymentStatusPaid,
			updatedPayment.TotalPrice,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order paid event")
		}

		res = dto.MapToPaymentResponse(updatedPayment)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
