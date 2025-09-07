// Package service provides business logic for order operations.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/redisutils"
)

// CreateOrderWithSaga creates an order with improved saga handling.
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

	var res *dto.OrderResponse

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		stateRepo := ds.SagaStateRepository()

		existingRes, shouldReturn, err := s.handleExistingOrder(ctx, req, orderRepo, stateRepo)
		if err != nil {
			return err
		}

		if shouldReturn {
			res = existingRes

			return nil
		}

		// Create new order if no existing order found
		orderItems := make([]entity.OrderItem, len(req.Items))
		for i := range req.Items {
			orderItems[i] = entity.OrderItem{
				ID:        uuid.New(),
				ProductID: req.Items[i].ProductID,
				Quantity:  req.Items[i].Quantity,
				UnitPrice: decimal.Zero, // Will be set by saga
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
		}

		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, "IDR", orderItems)
		if err != nil {
			return fmt.Errorf("failed to create order entity: %w", err)
		}

		if err := newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			s.logger.Errorf("failed to save order: %v", err)

			return fmt.Errorf("failed to save order: %w", err)
		}

		res = mapper.MapToOrderResponse(savedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.executeSagaWorkflow(ctx, res)
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
			return mapper.MapToOrderResponse(existingOrder), true, nil
		case constant.SagaStatusExecuting,
			constant.SagaStatusPending,
			constant.SagaStatusCompensating:
			return nil, false, fmt.Errorf("order is still being processed")
		case constant.SagaStatusFailed, constant.SagaStatusCompensated:
			s.logger.Infof("Retrying failed order %s", existingOrder.ID)
		}
	}

	return mapper.MapToOrderResponse(existingOrder), true, nil
}

// executeSagaWorkflow executes the saga based on configuration.
func (s *OrderService) executeSagaWorkflow(
	ctx context.Context,
	res *dto.OrderResponse,
) (*dto.OrderResponse, error) {
	if s.config.Saga.ExecutionMode == "sync" {
		return s.executeSagaSynchronously(ctx, res)
	}

	s.executeSagaAsynchronously(ctx, res.ID)
	res.Status = "processing"

	s.logger.Infof(
		"Your order is being processed. You will receive a confirmation once it's complete.",
	)

	return res, nil
}

// executeSagaSynchronously handles synchronous saga execution and error management.
func (s *OrderService) executeSagaSynchronously(
	ctx context.Context,
	res *dto.OrderResponse,
) (*dto.OrderResponse, error) {
	orderRepo := s.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, res.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	// Execute with timeout
	sagaCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := s.sagaOrchestrator.ExecuteOrderSaga(sagaCtx, order); err != nil {
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

// executeSagaAsynchronously executes saga in background.
func (s *OrderService) executeSagaAsynchronously(ctx context.Context, orderID uuid.UUID) {
	go func() {
		// Create background context with user authentication for async saga execution
		bgCtx := echoutils.PropagateUserContextToBackground(ctx)

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
