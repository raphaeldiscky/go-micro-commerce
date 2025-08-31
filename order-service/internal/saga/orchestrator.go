// Package saga provides saga coordination for order processing.
package saga

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// Orchestrator manages saga workflow execution.
type Orchestrator interface {
	ExecuteOrderSaga(ctx context.Context, order *entity.Order) error
}

// OrchestratorImpl implements the Orchestrator interface.
type OrchestratorImpl struct {
	orderSaga  *OrderSaga
	activities OrderActivities
	logger     logger.Logger
}

// NewSagaOrchestrator creates a new SagaOrchestrator instance.
func NewSagaOrchestrator(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer mq.KafkaProducerInterface,
	orderLifecycleProducer mq.KafkaProducerInterface,
	appLogger logger.Logger,
) Orchestrator {
	// Create activities
	activities := NewOrderActivities(
		dataStore,
		productClient,
		paymentRequestProducer,
		orderLifecycleProducer,
		appLogger,
	)

	// Create order saga
	orderSaga := NewOrderSaga(activities, appLogger)

	return &OrchestratorImpl{
		orderSaga:  orderSaga,
		activities: activities,
		logger:     appLogger,
	}
}

// ExecuteOrderSaga executes the order processing saga.
func (sc *OrchestratorImpl) ExecuteOrderSaga(ctx context.Context, order *entity.Order) error {
	sc.logger.Infof("Starting order saga execution for order: %s", order.ID)

	if err := sc.orderSaga.Execute(ctx, order); err != nil {
		sc.logger.Errorf("Order saga failed for order %s: %v", order.ID, err)

		return fmt.Errorf("saga execution failed: %w", err)
	}

	sc.logger.Infof("Order saga completed successfully for order: %s", order.ID)

	return nil
}
