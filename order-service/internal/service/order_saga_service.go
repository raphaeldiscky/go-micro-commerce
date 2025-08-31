// Package service provides business logic for order operations.
package service

import (
	"context"
	"fmt"
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

// CreateOrderWithSaga creates an order with improved saga handling.
func (s *OrderService) CreateOrderWithSaga(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	// Acquire distributed lock for idempotency
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
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	var res *dto.OrderResponse

	// Create order within transaction
	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		stateRepo := ds.SagaStateRepository()

		// Check for duplicate order using idempotency key
		existingOrder, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return err
		}

		if existingOrder != nil && existingOrder.CustomerID == req.CustomerID {
			// Check if saga already exists for this order
			sagaState, err := stateRepo.FindByOrderID(ctx, existingOrder.ID)
			if err != nil {
				return err
			}
			if sagaState != nil {
				switch sagaState.Status {
				case constant.SagaStatusCompleted:
					// Order already processed successfully
					res = dto.MapToOrderResponse(existingOrder)

					return nil
				case constant.SagaStatusExecuting, constant.SagaStatusPending:
					// Saga is still running
					return fmt.Errorf("order is still being processed")
				case constant.SagaStatusFailed, constant.SagaStatusCompensated:
					// Previous attempt failed, allow retry
					s.logger.Infof("Retrying failed order %s", existingOrder.ID)
				}
			}

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
				Price:     decimal.Zero, // Will be set by saga
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			orderItems = append(orderItems, orderItem)
		}

		// Create domain entity
		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, orderItems)
		if err != nil {
			return fmt.Errorf("failed to create order entity: %w", err)
		}

		// Set initial status
		if err := newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Save to repository
		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			return fmt.Errorf("failed to save order: %w", err)
		}

		res = dto.MapToOrderResponse(savedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Execute saga based on configuration
	if s.config.Saga.ExecutionMode == "sync" {
		// Synchronous execution - wait for saga to complete
		if err := s.executeSagaSynchronously(ctx, res.ID); err != nil {
			s.logger.Errorf("Synchronous saga execution failed: %v", err)
			// Update order status to failed
			s.UpdateOrderStatus(ctx, res.ID, constant.OrderStatusFailed)

			return nil, fmt.Errorf("order processing failed: %w", err)
		}

		// Retrieve updated order
		updatedOrder, err := s.GetOrder(ctx, res.ID)
		if err != nil {
			return nil, err
		}

		return updatedOrder, nil
	} else {
		// Asynchronous execution - return immediately
		s.executeSagaAsynchronously(ctx, res.ID)

		res.Status = "processing"

		s.logger.Infof("Your order is being processed. You will receive a confirmation once it's complete.")
	}

	return res, nil
}

// executeSagaSynchronously executes saga and waits for completion.
func (s *OrderService) executeSagaSynchronously(ctx context.Context, orderID uuid.UUID) error {
	orderRepo := s.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to retrieve order: %w", err)
	}

	// Execute with timeout
	sagaCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := s.sagaOrchestrator.ExecuteOrderSaga(sagaCtx, order); err != nil {
		return fmt.Errorf("saga execution failed: %w", err)
	}

	return nil
}

// executeSagaAsynchronously executes saga in background.
func (s *OrderService) executeSagaAsynchronously(ctx context.Context, orderID uuid.UUID) {
	go func() {
		// Create background context with values from original context
		bgCtx := context.Background()

		// Copy trace ID if present
		if traceID := ctx.Value("trace_id"); traceID != nil {
			bgCtx = context.WithValue(bgCtx, "trace_id", traceID)
		}

		// Add timeout
		sagaCtx, cancel := context.WithTimeout(bgCtx, 30*time.Minute)
		defer cancel()

		orderRepo := s.dataStore.OrderRepository()

		order, err := orderRepo.FindByID(sagaCtx, orderID)
		if err != nil {
			s.logger.Errorf("Failed to retrieve order for saga: %v", err)
			s.handleSagaError(orderID, err)

			return
		}

		s.logger.Infof("Starting async saga for order %s", orderID)

		if err := s.sagaOrchestrator.ExecuteOrderSaga(sagaCtx, order); err != nil {
			s.logger.Errorf("Async saga failed for order %s: %v", orderID, err)
			s.handleSagaError(orderID, err)
		} else {
			s.logger.Infof("Async saga completed for order %s", orderID)
		}
	}()
}

// handleSagaError handles saga execution errors.
func (s *OrderService) handleSagaError(orderID uuid.UUID, err error) {
	ctx := context.Background()

	// Update order status to failed
	updateOrder, updateErr := s.UpdateOrderStatus(ctx, orderID, constant.OrderStatusFailed)
	if updateErr != nil {
		s.logger.Errorf("Failed to update order status: %v", updateErr)

		return
	}

	// Send failure notification
	s.NotifyOrderFailure(ctx, orderID, updateOrder.Status, err.Error())
}

// GetSagaStatus retrieves the current saga execution status.
func (s *OrderService) GetSagaStatus(
	ctx context.Context,
	orderID uuid.UUID,
) (*dto.SagaStatusResponse, error) {
	stateRepo := s.dataStore.SagaStateRepository()

	sagaState, err := stateRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saga status: %w", err)
	}

	if sagaState == nil {
		return nil, fmt.Errorf("no saga found for order %s", orderID)
	}

	return &dto.SagaStatusResponse{
		SagaID:           sagaState.ID,
		OrderID:          sagaState.OrderID,
		Status:           sagaState.Status,
		CurrentStep:      sagaState.CurrentStep,
		ExecutedSteps:    sagaState.ExecutedSteps,
		CompensatedSteps: sagaState.CompensatedSteps,
		Error:            sagaState.Error,
		CreatedAt:        sagaState.CreatedAt,
		UpdatedAt:        sagaState.UpdatedAt,
		CompletedAt:      sagaState.CompletedAt,
	}, nil
}
