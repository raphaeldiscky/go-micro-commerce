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
	lock, err := s.acquireOrderLock(ctx, req.IdempotencyKey, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer s.releaseOrderLock(ctx, lock)

	res, err := s.createOrderWithinTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	return s.executeSagaWorkflow(ctx, res)
}

// acquireOrderLock acquires a distributed lock for order creation idempotency.
func (s *OrderService) acquireOrderLock(
	ctx context.Context,
	idempotencyKey uuid.UUID,
	customerID uuid.UUID,
) (*redislock.Lock, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(idempotencyKey, customerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	return lockRepo.Get(ctx, lockKey, ttl, opt)
}

// releaseOrderLock releases the distributed lock.
func (s *OrderService) releaseOrderLock(ctx context.Context, lock *redislock.Lock) {
	lockRepo := s.dataStore.LockRepository()
	if err := lockRepo.Release(ctx, lock); err != nil {
		s.logger.Warnf("failed to release lock: %v", err)
	}
}

// createOrderWithinTransaction handles order creation within a transaction.
func (s *OrderService) createOrderWithinTransaction(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	var res *dto.OrderResponse

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		stateRepo := ds.SagaStateRepository()

		// Check for existing order and handle accordingly
		existingRes, shouldReturn, err := s.handleExistingOrder(ctx, req, orderRepo, stateRepo)
		if err != nil {
			return err
		}

		if shouldReturn {
			res = existingRes

			return nil
		}

		// Create new order if no existing order found
		newRes, err := s.createNewOrder(ctx, req, orderRepo)
		if err != nil {
			return err
		}

		res = newRes

		return nil
	})

	return res, err
}

// handleExistingOrder checks for duplicate orders and handles saga state.
func (s *OrderService) handleExistingOrder(
	ctx context.Context,
	req dto.CreateOrderRequest,
	orderRepo repository.OrderRepositoryInterface,
	stateRepo repository.SagaStateRepositoryInterface,
) (*dto.OrderResponse, bool, error) {
	existingOrder, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, false, err
	}

	if existingOrder == nil || existingOrder.CustomerID != req.CustomerID {
		return nil, false, nil // No existing order, continue with creation
	}

	sagaState, err := stateRepo.FindByOrderID(ctx, existingOrder.ID)
	if err != nil {
		return nil, false, err
	}

	if sagaState != nil {
		switch sagaState.Status {
		case constant.SagaStatusCompleted:
			return dto.MapToOrderResponse(existingOrder), true, nil
		case constant.SagaStatusExecuting,
			constant.SagaStatusPending,
			constant.SagaStatusCompensating:
			return nil, false, fmt.Errorf("order is still being processed")
		case constant.SagaStatusFailed, constant.SagaStatusCompensated:
			s.logger.Infof("Retrying failed order %s", existingOrder.ID)
		}
	}

	return dto.MapToOrderResponse(existingOrder), true, nil
}

// createNewOrder creates a new order entity and saves it.
func (s *OrderService) createNewOrder(
	ctx context.Context,
	req dto.CreateOrderRequest,
	orderRepo repository.OrderRepositoryInterface,
) (*dto.OrderResponse, error) {
	orderItems := s.buildOrderItems(req.Items)

	newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, orderItems)
	if err != nil {
		return nil, fmt.Errorf("failed to create order entity: %w", err)
	}

	if err := newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	savedOrder, err := orderRepo.Create(ctx, newOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return dto.MapToOrderResponse(savedOrder), nil
}

// buildOrderItems creates order items from the request.
func (s *OrderService) buildOrderItems(items []dto.CreateOrderItemRequest) []entity.OrderItem {
	orderItems := make([]entity.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = entity.OrderItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     decimal.Zero, // Will be set by saga
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
	}

	return orderItems
}

// executeSagaWorkflow executes the saga based on configuration.
func (s *OrderService) executeSagaWorkflow(
	ctx context.Context,
	res *dto.OrderResponse,
) (*dto.OrderResponse, error) {
	if s.config.Saga.ExecutionMode == "sync" {
		return s.handleSyncSagaExecution(ctx, res)
	}

	s.executeSagaAsynchronously(ctx, res.ID)
	res.Status = "processing"

	s.logger.Infof(
		"Your order is being processed. You will receive a confirmation once it's complete.",
	)

	return res, nil
}

// handleSyncSagaExecution handles synchronous saga execution and error management.
func (s *OrderService) handleSyncSagaExecution(
	ctx context.Context,
	res *dto.OrderResponse,
) (*dto.OrderResponse, error) {
	if err := s.executeSagaSynchronously(ctx, res.ID); err != nil {
		s.logger.Errorf("Synchronous saga execution failed: %v", err)
		// Update order status to failed
		order, updateErr := s.UpdateOrderStatus(ctx, res.ID, constant.OrderStatusFailed)
		if updateErr != nil {
			s.logger.Errorf("Failed to update order status: OrderID %v, %v", order.ID, updateErr)
		}

		return nil, fmt.Errorf("order processing failed: OrderID %v, %w", order.ID, err)
	}

	// Retrieve updated order
	updatedOrder, err := s.GetOrder(ctx, res.ID)
	if err != nil {
		return nil, err
	}

	return updatedOrder, nil
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
		if traceID := ctx.Value(constant.CtxTraceIDKey); traceID != nil {
			bgCtx = context.WithValue(bgCtx, constant.CtxTraceIDKey, traceID)
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
	err = s.NotifyOrderFailure(ctx, orderID, updateOrder.Status, err.Error())
	if err != nil {
		s.logger.Errorf("Failed to send order failure notification: %v", err)
	}
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
