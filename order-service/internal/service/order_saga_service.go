// Package service provides business logic for order operations.
package service

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/redisutils"
)

// CreateOrderWithSaga creates an order and processes it using the saga pattern.
func (s *OrderService) CreateOrderWithSaga(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
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

	res := new(dto.OrderResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check for duplicate order using idempotency key
		existingOrder, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return err
		}

		if existingOrder != nil && existingOrder.CustomerID == req.CustomerID {
			res = dto.MapToOrderResponse(existingOrder)

			return nil
		}

		// Create order items from request
		var orderItems []entity.OrderItem

		for _, item := range req.Items {
			orderItem := entity.OrderItem{
				ID:        uuid.New(),
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     decimal.Zero, // Will be set by saga when products are validated
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			orderItems = append(orderItems, orderItem)
		}

		// Create domain entity
		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, orderItems)
		if err != nil {
			return err
		}

		// Set initial status for saga processing
		if err := newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
			return err
		}

		// Save to repository
		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			return err
		}

		res = dto.MapToOrderResponse(savedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Execute saga workflow asynchronously
	go func() {
		sagaCtx := context.Background() // Use background context for async execution

		if s.sagaOrchestrator == nil {
			s.logger.Warnf(
				"Saga orchestrator is not available, order %s will remain in pending status",
				res.ID,
			)

			return
		}

		// Get the fresh order for saga processing
		orderRepo := s.dataStore.OrderRepository()

		order, err := orderRepo.FindByID(sagaCtx, res.ID)
		if err != nil {
			s.logger.Errorf("Failed to get order for saga processing: %v", err)

			return
		}

		sagaErr := s.sagaOrchestrator.ExecuteOrderSaga(sagaCtx, order)
		if sagaErr == nil {
			return
		}

		s.logger.Errorf("Saga execution failed for order %s: %v", res.ID, sagaErr)
		// Update order status to canceled on saga failure
		if _, updateErr := s.UpdateOrderStatus(sagaCtx, res.ID, constant.OrderStatusCanceled); updateErr != nil {
			s.logger.Errorf("Failed to update order status to canceled: %v", updateErr)
		}
	}()

	return res, nil
}
