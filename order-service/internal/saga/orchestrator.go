// Package saga provides saga coordination for order processing.
package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// Orchestrator manages saga workflow execution with state persistence.
type Orchestrator struct {
	executor              *Executor
	orderSaga             *OrderSaga
	dataStore             repository.DataStore
	logger                logger.Logger
	asyncExecutionTimeout time.Duration
}

// NewSagaOrchestrator creates a new  orchestrator.
func NewSagaOrchestrator(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer mq.KafkaProducerInterface,
	orderLifecycleProducer mq.KafkaProducerInterface,
	appLogger logger.Logger,
) Orchestrator {
	// Create executor
	executor := NewExecutor(dataStore, appLogger)

	// Create activities
	activities := NewOrderActivities(
		dataStore,
		productClient,
		paymentRequestProducer,
		orderLifecycleProducer,
		appLogger,
	)

	// Create  order saga
	orderSaga := NewOrderSaga(activities, appLogger)

	// Configure saga steps in executor
	orderSaga.ConfigureSteps(executor)

	return Orchestrator{
		executor:              executor,
		orderSaga:             orderSaga,
		dataStore:             dataStore,
		logger:                appLogger,
		asyncExecutionTimeout: 30 * time.Minute,
	}
}

// ExecuteOrderSaga executes the order processing saga with proper async handling.
func (o *Orchestrator) ExecuteOrderSaga(ctx context.Context, order *entity.Order) error {
	o.logger.Infof("Starting order saga execution for order: %s", order.ID)

	// Create a context with timeout for async execution
	sagaCtx, cancel := context.WithTimeout(ctx, o.asyncExecutionTimeout)
	defer cancel()

	// Execute the saga
	if err := o.executor.Execute(sagaCtx, order); err != nil {
		o.logger.Errorf("Order saga failed for order %s: %v", order.ID, err)

		return fmt.Errorf("saga execution failed: %w", err)
	}

	o.logger.Infof("Order saga completed successfully for order: %s", order.ID)

	return nil
}

// ExecuteOrderSagaAsync executes the saga asynchronously with proper tracking.
func (o *Orchestrator) ExecuteOrderSagaAsync(
	ctx context.Context,
	order *entity.Order,
) {
	// Create a derived context that inherits values but not cancellation
	sagaCtx := context.WithValue(context.Background(), "order_id", order.ID)
	sagaCtx = context.WithValue(sagaCtx, "trace_id", ctx.Value("trace_id"))

	// Add timeout
	sagaCtx, cancel := context.WithTimeout(sagaCtx, o.asyncExecutionTimeout)

	go func() {
		defer cancel()

		o.logger.Infof("Starting async saga execution for order: %s", order.ID)

		if err := o.executor.Execute(sagaCtx, order); err != nil {
			o.logger.Errorf("Async saga execution failed for order %s: %v", order.ID, err)

			// Update order status to failed
			// Note: In production, this should be done through proper event handling
			o.handleSagaFailure(order.ID, err)
		} else {
			o.logger.Infof("Async saga execution completed for order: %s", order.ID)
		}
	}()
}

// RecoverFailedSagas recovers and retries failed sagas.
func (o *Orchestrator) RecoverFailedSagas(ctx context.Context) error {
	o.logger.Info("Starting recovery of failed sagas")
	stateRepo := o.dataStore.SagaStateRepository()
	// Find failed or pending sagas
	failedSagas, err := stateRepo.FindPendingOrFailed(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to find sagas for recovery: %w", err)
	}

	o.logger.Infof("Found %d sagas to recover", len(failedSagas))

	for _, sagaState := range failedSagas {
		o.logger.Infof("Recovering saga %s for order %s", sagaState.ID, sagaState.OrderID)

		// Retrieve the order
		// Note: You'll need to inject order repository or service
		order, err := o.getOrder(ctx, sagaState.OrderID)
		if err != nil {
			o.logger.Errorf("Failed to retrieve order %s: %v", sagaState.OrderID, err)

			continue
		}

		// Retry the saga execution
		go func(order *entity.Order) {
			recoveryCtx, cancel := context.WithTimeout(
				context.Background(),
				o.asyncExecutionTimeout,
			)
			defer cancel()

			if err := o.executor.Execute(recoveryCtx, order); err != nil {
				o.logger.Errorf("Failed to recover saga for order %s: %v", order.ID, err)
			} else {
				o.logger.Infof("Successfully recovered saga for order %s", order.ID)
			}
		}(order)
	}

	return nil
}

// handleSagaFailure handles saga failure by updating order status.
func (o *Orchestrator) handleSagaFailure(orderID uuid.UUID, err error) {
	// This should be implemented based on your order service interface
	o.logger.Errorf("Handling saga failure for order %s: %v", orderID, err)
	// Update order status to failed/canceled
}

// getOrder retrieves an order (placeholder - implement based on your architecture).
func (o *Orchestrator) getOrder(ctx context.Context, orderID uuid.UUID) (*entity.Order, error) {
	// This should be implemented to retrieve the order from repository
	// For now, returning an error as placeholder
	return nil, fmt.Errorf("getOrder not implemented")
}
