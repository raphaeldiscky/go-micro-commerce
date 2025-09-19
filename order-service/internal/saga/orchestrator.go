// Package saga provides saga coordination for order processing.
package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// Orchestrator defines the interface for saga orchestration.
type Orchestrator interface {
	ExecuteOrderSaga(ctx context.Context, payload *Payload) error
	ExecuteOrderSagaAsync(ctx context.Context, payload *Payload)
	RecoverFailedSagas(ctx context.Context) error
	TriggerSagaCompensation(ctx context.Context, orderID uuid.UUID) error
}

// orchestrator manages saga workflow execution with state persistence.
type orchestrator struct {
	executor         *Executor
	dataStore        repository.DataStore
	logger           logger.Logger
	executionTimeout time.Duration
}

// NewSagaOrchestrator creates a new orchestrator.
func NewSagaOrchestrator(
	dataStore repository.DataStore,
	productClient client.ProductClient,
	fulfillmentClient client.FulfillmentClient,
	paymentClient client.PaymentClient,
	asynqClient asynq.Client,
	taskCancellationService asynq.TaskCancellationService,
	appLogger logger.Logger,
	cfg *config.Config,
) Orchestrator {
	// Create executor
	executor := NewExecutor(dataStore, cfg, appLogger)

	// Create activities
	activities := NewOrderActivities(
		dataStore,
		productClient,
		fulfillmentClient,
		paymentClient,
		asynqClient,
		taskCancellationService,
		appLogger,
	)

	// Create  order saga
	orderSaga := NewOrderSaga(activities)

	// Configure saga steps in executor
	orderSaga.ConfigureSteps(executor)

	return &orchestrator{
		executor:         executor,
		dataStore:        dataStore,
		logger:           appLogger,
		executionTimeout: cfg.Saga.DefaultExecutionTimeout,
	}
}

// ExecuteOrderSaga executes the order processing saga with proper async handling.
func (o *orchestrator) ExecuteOrderSaga(ctx context.Context, payload *Payload) error {
	o.logger.Infof("Starting order saga execution for order: %s", payload.Order.ID)

	// Check if saga is already in compensating state - if so, don't try to extract auth from context
	// since compensation uses auth stored in saga metadata
	sagaRepo := o.dataStore.SagaStateRepository()
	sagaState, stateErr := sagaRepo.FindByOrderID(ctx, payload.Order.ID)
	isCompensating := stateErr == nil && sagaState.Status == constant.SagaStatusCompensating

	// Create execution context with auth if needed
	sagaCtx, err := o.createSagaExecutionContext(ctx, !isCompensating)
	if err != nil {
		return fmt.Errorf("failed to create saga context: %w", err)
	}

	defer func() {
		if cancel, ok := sagaCtx.Value("cancel").(context.CancelFunc); ok {
			cancel()
		}
	}()

	// Execute the saga
	if execErr := o.executor.Execute(sagaCtx, payload); execErr != nil {
		o.logger.Errorf("Order saga failed for order %s: %v", payload.Order.ID, execErr)

		return fmt.Errorf("saga execution failed: %w", execErr)
	}

	o.logger.Infof("Order saga completed successfully for order: %s", payload.Order.ID)

	return nil
}

// ExecuteOrderSagaAsync executes the saga asynchronously with proper tracking.
func (o *orchestrator) ExecuteOrderSagaAsync(
	ctx context.Context,
	payload *Payload,
) {
	go func() {
		o.logger.Infof("Starting async saga execution for order: %s", payload.Order.ID)

		// Create async execution context
		sagaCtx, err := o.createAsyncSagaExecutionContext(ctx, payload.Order.ID)
		if err != nil {
			o.logger.Errorf(
				"Failed to create async saga context for order %s: %v",
				payload.Order.ID,
				err,
			)

			return
		}

		defer func() {
			if cancel, ok := sagaCtx.Value("cancel").(context.CancelFunc); ok {
				cancel()
			}
		}()

		if execErr := o.executor.Execute(sagaCtx, payload); execErr != nil {
			o.logger.Errorf(
				"Async saga execution failed for order %s: %v",
				payload.Order.ID,
				execErr,
			)

			// Update order status to failed
			// Note: In production, this should be done through proper event handling
			o.handleSagaFailure(payload.Order.ID, execErr)
		} else {
			o.logger.Infof("Async saga execution completed for order: %s", payload.Order.ID)
		}
	}()
}

// RecoverFailedSagas recovers and retries failed sagas.
func (o *orchestrator) RecoverFailedSagas(ctx context.Context) error {
	o.logger.Info("Starting recovery of failed sagas")
	stateRepo := o.dataStore.SagaStateRepository()
	// Find failed or pending sagas
	failedSagas, err := stateRepo.FindPendingOrFailed(ctx, pkgconstant.DefaultMaxLimit)
	if err != nil {
		return fmt.Errorf("failed to find sagas for recovery: %w", err)
	}

	o.logger.Infof("Found %d sagas to recover", len(failedSagas))

	for _, sagaState := range failedSagas {
		o.logger.Infof("Recovering saga %s for order %s", sagaState.ID, sagaState.OrderID)

		orderRepo := o.dataStore.OrderRepository()

		order, rowErr := orderRepo.FindByID(ctx, sagaState.OrderID)
		if rowErr != nil {
			if rowErr.Error() == constant.OrderNotFoundErrorMessage {
				o.logger.Errorf("Order not found for saga recovery: %s", sagaState.OrderID)
			} else {
				o.logger.Errorf("Failed to retrieve order %s: %v", sagaState.OrderID, rowErr)
			}

			continue
		}

		// Retry the saga execution
		go func(order *entity.Order, sagaState *entity.SagaState) {
			// Use parent context instead of background context to preserve any available context values
			// The executor will retrieve user auth from stored saga metadata
			recoveryCtx, cancel := context.WithTimeout(
				ctx,
				o.executionTimeout,
			)
			defer cancel()

			// Extract shipping data from saga state if available
			var shipping dto.Shipping

			if shippingData, exists := sagaState.Data["shipping"]; exists {
				if shippingMap, ok := shippingData.(map[string]any); ok {
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

			if err = o.executor.Execute(recoveryCtx, payload); err != nil {
				o.logger.Errorf("Failed to recover saga for order %s: %v", payload.Order.ID, err)
			} else {
				o.logger.Infof("Successfully recovered saga for order %s", payload.Order.ID)
			}
		}(order, sagaState)
	}

	return nil
}

// handleSagaFailure handles saga failure by updating order status.
func (o *orchestrator) handleSagaFailure(orderID uuid.UUID, err error) {
	// This should be implemented based on your order service interface
	o.logger.Errorf("Handling saga failure for order %s: %v", orderID, err)
	// Update order status to failed/canceled
}

// convertMapToShippingRequest converts a map from saga state back to ShippingRequest.
// This is a simplified implementation - in production consider using JSON marshal/unmarshal.
func convertMapToShippingRequest(data map[string]any) dto.Shipping {
	var shipping dto.Shipping

	if carrierID, ok := data["carrier_id"].(string); ok {
		shipping.CarrierID = carrierID
	}

	// Add more field conversions as needed based on your ShippingRequest structure
	// This is a basic implementation - you might want to use JSON marshal/unmarshal for robustness

	return shipping
}

// createSagaExecutionContext creates a context for saga execution with timeout and optional auth.
func (o *orchestrator) createSagaExecutionContext(
	ctx context.Context,
	includeAuth bool,
) (context.Context, error) {
	// Create a context with timeout for saga execution
	sagaCtx, cancel := context.WithTimeout(ctx, o.executionTimeout)

	// Store cancel function in context for cleanup
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxCancel, cancel)

	// Only extract and set user auth from context if requested
	if includeAuth {
		userAuth, err := echoutils.GetUserAuthContexts(ctx)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to get user auth: %w", err)
		}

		// Add user auth to context for the first step to pick up
		sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxUserID, userAuth.UserID)
		sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxEmail, userAuth.Email)
		sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxRoles, userAuth.Roles)
		sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxIsActive, userAuth.IsActive)
	}

	return sagaCtx, nil
}

// createAsyncSagaExecutionContext creates a context for async saga execution.
func (o *orchestrator) createAsyncSagaExecutionContext(
	ctx context.Context,
	orderID uuid.UUID,
) (context.Context, error) {
	// Create a derived context with user authentication for async saga execution
	sagaCtx := echoutils.PropagateUserContextToBackground(ctx)
	sagaCtx = context.WithValue(sagaCtx, constant.CtxOrderIDKey, orderID)
	sagaCtx = context.WithValue(sagaCtx, constant.CtxTraceIDKey, ctx.Value(constant.CtxTraceIDKey))

	// Add timeout
	sagaCtx, cancel := context.WithTimeout(sagaCtx, o.executionTimeout)

	// Store cancel function in context for cleanup
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxCancel, cancel)

	// Extract user auth from original context for async execution
	userAuth, err := echoutils.GetUserAuthContexts(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get user auth for async saga: %w", err)
	}

	// Add user auth to saga context for the first step to pick up
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxUserID, userAuth.UserID)
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxEmail, userAuth.Email)
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxRoles, userAuth.Roles)
	sagaCtx = context.WithValue(sagaCtx, pkgconstant.CtxIsActive, userAuth.IsActive)

	return sagaCtx, nil
}

// TriggerSagaCompensation triggers immediate compensation for the saga associated with the given order.
func (o *orchestrator) TriggerSagaCompensation(
	ctx context.Context,
	orderID uuid.UUID,
) error {
	sagaRepo := o.dataStore.SagaStateRepository()
	orderRepo := o.dataStore.OrderRepository()

	// Find the saga state for this order
	sagaState, err := sagaRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		if err.Error() == constant.SagaStateNotFoundErrorMessage {
			o.logger.Warnf("No saga state found for order %s, skipping compensation", orderID)
			return nil // Not an error - order might not have been processed through saga
		}

		return fmt.Errorf("failed to find saga state for order %s: %w", orderID, err)
	}

	// Check if saga is in a state that can be compensated
	switch sagaState.Status {
	case constant.SagaStatusCompleted:
		o.logger.Infof("Order %s saga already completed, no compensation needed", orderID)
		return nil
	case constant.SagaStatusCompensated:
		o.logger.Infof("Order %s saga already compensated", orderID)
		return nil
	case constant.SagaStatusFailed:
		o.logger.Infof("Order %s saga already failed", orderID)
		return nil
	case constant.SagaStatusCompensating:
		o.logger.Infof("Order %s saga already compensating", orderID)
		return nil
	case constant.SagaStatusPending, constant.SagaStatusExecuting:
		// These are the states where we can trigger compensation
		o.logger.Infof(
			"Order %s saga is in %s state, triggering immediate compensation",
			orderID,
			sagaState.Status,
		)
	}

	// Get the order for saga payload
	order, err := orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to find order for compensation: %w", err)
	}

	// Create saga payload for compensation
	payload := &Payload{
		Order: order,
		// Note: We don't have shipping data in this context, but compensation doesn't need it
		Shipping: dto.Shipping{},
	}

	// Mark saga as compensating first
	err = sagaRepo.MarkAsCompensating(ctx, sagaState.ID)
	if err != nil {
		return fmt.Errorf("failed to mark saga as compensating for order %s: %w", orderID, err)
	}

	// Execute compensation immediately using async context with auth
	go func() {
		// Start with the incoming context to preserve auth information
		compensationCtx := echoutils.PropagateUserContextToBackground(ctx)
		// Add timeout
		compensationCtx, cancel := context.WithTimeout(compensationCtx, o.executionTimeout)
		defer cancel()

		// Ensure important context values are propagated
		compensationCtx = context.WithValue(compensationCtx, constant.CtxOrderIDKey, orderID)
		if traceID := ctx.Value(constant.CtxTraceIDKey); traceID != nil {
			compensationCtx = context.WithValue(compensationCtx, constant.CtxTraceIDKey, traceID)
		}

		o.logger.Infof(
			"Starting immediate compensation for order %s (saga ID: %s)",
			orderID,
			sagaState.ID,
		)

		// Use the orchestrator's existing saga execution which will handle compensation
		// since the saga is already marked as compensating
		if compensationErr := o.ExecuteOrderSaga(compensationCtx, payload); compensationErr != nil {
			o.logger.Errorf(
				"Failed to execute saga compensation for order %s: %v",
				orderID,
				compensationErr,
			)
		} else {
			o.logger.Infof("Successfully completed saga compensation for order %s", orderID)
		}
	}()

	o.logger.Infof(
		"Successfully triggered immediate saga compensation for order %s (saga ID: %s)",
		orderID,
		sagaState.ID,
	)

	return nil
}
