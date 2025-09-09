// Package saga provides saga coordination for order processing.
package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
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
	paymentRequestProducer kafka.ProducerInterface,
	orderLifecycleProducer kafka.ProducerInterface,
	fulfillmentRequestProducer kafka.ProducerInterface,
	fulfillmentClient client.FulfillmentClientInterface,
	paymentClient client.PaymentClientInterface,
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
		fulfillmentRequestProducer,
		fulfillmentClient,
		paymentClient,
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
func (o *Orchestrator) ExecuteOrderSaga(ctx context.Context, payload *Payload) error {
	o.logger.Infof("Starting order saga execution for order: %s", payload.Order.ID)

	// Create a context with timeout for async execution
	sagaCtx, cancel := context.WithTimeout(ctx, o.asyncExecutionTimeout)
	defer cancel()

	// Execute the saga
	if err := o.executor.Execute(sagaCtx, payload); err != nil {
		o.logger.Errorf("Order saga failed for order %s: %v", payload.Order.ID, err)

		return fmt.Errorf("saga execution failed: %w", err)
	}

	o.logger.Infof("Order saga completed successfully for order: %s", payload.Order.ID)

	return nil
}

// ExecuteOrderSagaAsync executes the saga asynchronously with proper tracking.
func (o *Orchestrator) ExecuteOrderSagaAsync(
	ctx context.Context,
	payload *Payload,
) {
	// Create a derived context with user authentication for async saga execution
	sagaCtx := echoutils.PropagateUserContextToBackground(ctx)
	sagaCtx = context.WithValue(sagaCtx, constant.CtxOrderIDKey, payload.Order.ID)
	sagaCtx = context.WithValue(sagaCtx, constant.CtxTraceIDKey, ctx.Value(constant.CtxTraceIDKey))

	// Add timeout
	sagaCtx, cancel := context.WithTimeout(sagaCtx, o.asyncExecutionTimeout)

	go func() {
		defer cancel()

		o.logger.Infof("Starting async saga execution for order: %s", payload.Order.ID)

		if err := o.executor.Execute(sagaCtx, payload); err != nil {
			o.logger.Errorf("Async saga execution failed for order %s: %v", payload.Order.ID, err)

			// Update order status to failed
			// Note: In production, this should be done through proper event handling
			o.handleSagaFailure(payload.Order.ID, err)
		} else {
			o.logger.Infof("Async saga execution completed for order: %s", payload.Order.ID)
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
		orderRepo := o.dataStore.OrderRepository()

		order, err := orderRepo.FindByID(ctx, sagaState.OrderID)
		if err != nil {
			o.logger.Errorf("Failed to retrieve order %s: %v", sagaState.OrderID, err)

			continue
		}

		// Retry the saga execution
		go func(order *entity.Order, sagaState *entity.SagaState) {
			recoveryCtx, cancel := context.WithTimeout(
				context.Background(),
				o.asyncExecutionTimeout,
			)
			defer cancel()

			// Extract shipping data from saga state if available
			var shipping dto.Shipping

			if shippingData, exists := sagaState.Data["shipping_request"]; exists {
				if shippingMap, ok := shippingData.(map[string]interface{}); ok {
					// Convert map to ShippingRequest - this is a simplified approach
					// In production, you might want to use JSON marshal/unmarshal for type safety
					o.logger.Infof("Recovered shipping data from saga state for order %s", order.ID)

					shipping = convertMapToShippingRequest(shippingMap)
				}
			} else {
				// If no shipping data in saga state, we'll proceed with empty shipping
				// This handles cases where saga failed before Step 1 completed
				o.logger.Warnf("No shipping data found in saga state for order %s during recovery", order.ID)
			}

			payload := &Payload{
				Order:    order,
				Shipping: shipping,
			}

			if err := o.executor.Execute(recoveryCtx, payload); err != nil {
				o.logger.Errorf("Failed to recover saga for order %s: %v", payload.Order.ID, err)
			} else {
				o.logger.Infof("Successfully recovered saga for order %s", payload.Order.ID)
			}
		}(order, sagaState)
	}

	return nil
}

// handleSagaFailure handles saga failure by updating order status.
func (o *Orchestrator) handleSagaFailure(orderID uuid.UUID, err error) {
	// This should be implemented based on your order service interface
	o.logger.Errorf("Handling saga failure for order %s: %v", orderID, err)
	// Update order status to failed/canceled
}

// convertMapToShippingRequest converts a map from saga state back to ShippingRequest.
// This is a simplified implementation - in production consider using JSON marshal/unmarshal.
func convertMapToShippingRequest(data map[string]interface{}) dto.Shipping {
	var shipping dto.Shipping

	if carrierID, ok := data["carrier_id"].(string); ok {
		shipping.CarrierID = carrierID
	}

	// Add more field conversions as needed based on your ShippingRequest structure
	// This is a basic implementation - you might want to use JSON marshal/unmarshal for robustness

	return shipping
}
