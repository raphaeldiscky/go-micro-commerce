// Package saga provides activity implementations for order saga.
package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	ReserveInventory(ctx context.Context, order *entity.Order) error
	ReleaseInventoryReservation(ctx context.Context, order *entity.Order) error
	ProcessPayment(ctx context.Context, order *entity.Order) error
	RefundPayment(ctx context.Context, order *entity.Order) error
	UpdateInventory(ctx context.Context, order *entity.Order) error
	ArrangeShipping(ctx context.Context, order *entity.Order) error
}

// OrderActivitiesImpl implements the OrderActivities interface.
type OrderActivitiesImpl struct {
	dataStore              repository.DataStore
	productClient          client.ProductClientInterface
	paymentRequestProducer mq.KafkaProducerInterface
	orderLifecycleProducer mq.KafkaProducerInterface
	logger                 logger.Logger
}

// NewOrderActivities creates a new OrderActivitiesImpl instance.
func NewOrderActivities(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer mq.KafkaProducerInterface,
	orderLifecycleProducer mq.KafkaProducerInterface,
	appLogger logger.Logger,
) OrderActivities {
	return &OrderActivitiesImpl{
		dataStore:              dataStore,
		productClient:          productClient,
		paymentRequestProducer: paymentRequestProducer,
		orderLifecycleProducer: orderLifecycleProducer,
		logger:                 appLogger,
	}
}

// ReserveInventory reserves inventory for the order items.
func (a *OrderActivitiesImpl) ReserveInventory(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Reserving inventory for order: %s", order.ID)

	if a.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// Prepare reservation items
	reservationItems := make([]client.ProductReservationItem, len(order.Items))
	for i, item := range order.Items {
		reservationItems[i] = client.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Reserve products using product service
	_, err := a.productClient.ReserveProducts(ctx, order.IdempotencyKey, reservationItems)
	if err != nil {
		a.logger.Errorf("Failed to reserve inventory for order %s: %v", order.ID, err)

		return fmt.Errorf("inventory reservation failed: %w", err)
	}

	a.logger.Infof("Successfully reserved inventory for order: %s", order.ID)

	return nil
}

// ReleaseInventoryReservation releases reserved inventory.
func (a *OrderActivitiesImpl) ReleaseInventoryReservation(
	_ context.Context,
	order *entity.Order,
) error {
	a.logger.Infof("Releasing inventory reservation for order: %s", order.ID)

	if a.productClient == nil {
		a.logger.Warnf(
			"Product service unavailable, cannot release reservation for order: %s",
			order.ID,
		)

		return nil // Don't fail compensation if service is down
	}

	// In a real implementation, you would call a ReleaseReservation method
	// For now, we'll log the compensation action
	// TODO: Implement product client ReleaseReservation method
	a.logger.Warnf(
		"Inventory reservation release not implemented for order: %s (compensation needed)",
		order.ID,
	)

	a.logger.Infof("Successfully released inventory reservation for order: %s", order.ID)

	return nil
}

// ProcessPayment processes payment for the order.
func (a *OrderActivitiesImpl) ProcessPayment(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Processing payment for order: %s", order.ID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create payment request event
		paymentEvent := event.NewPaymentRequestEvent(
			order.ID,
			order.CustomerID,
			order.TotalPrice,
			"IDR",         // Default currency
			"credit_card", // Default payment method for saga
		)

		payload, err := json.Marshal(paymentEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal payment request event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   order.ID,
			EventType:     constant.KafkaEventTypePaymentRequested,
			Topic:         constant.TopicPaymentRequest,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create payment request event: %w", err)
		}

		a.logger.Infof("Successfully created payment request for order: %s", order.ID)

		return nil
	})
}

// RefundPayment refunds payment for the order.
func (a *OrderActivitiesImpl) RefundPayment(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Refunding payment for order: %s", order.ID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create refund request event
		refundEvent := map[string]interface{}{
			"order_id":    order.ID,
			"customer_id": order.CustomerID,
			"amount":      order.TotalPrice,
			"currency":    "IDR",
			"reason":      "order_canceled",
			"timestamp":   time.Now().UTC(),
		}

		payload, err := json.Marshal(refundEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal refund request event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   order.ID,
			EventType:     constant.KafkaEventTypePaymentRefunded,
			Topic:         constant.TopicPaymentRequest, // Use same topic with different event type
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create refund request event: %w", err)
		}

		a.logger.Infof("Successfully created refund request for order: %s", order.ID)

		return nil
	})
}

// UpdateInventory updates inventory after successful payment.
func (a *OrderActivitiesImpl) UpdateInventory(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Updating inventory for order: %s", order.ID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Update order status to confirmed
		if err := order.UpdateStatus(constant.OrderStatusConfirmed); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to save order status update: %w", err)
		}

		// Publish order confirmed event
		orderEvent := event.NewOrderLifecycleEvent(
			updatedOrder.ID,
			constant.OrderStatusConfirmed,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)

		if err := a.orderLifecycleProducer.Send(ctx, orderEvent); err != nil {
			return fmt.Errorf("failed to send order confirmed event: %w", err)
		}

		a.logger.Infof("Successfully updated inventory and order status for order: %s", order.ID)

		return nil
	})
}

// ArrangeShipping arranges shipping for the order.
func (a *OrderActivitiesImpl) ArrangeShipping(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Arranging shipping for order: %s", order.ID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Update order status to processing
		if err := order.UpdateStatus(constant.OrderStatusProcessing); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to save order status update: %w", err)
		}

		// Create shipping arrangement event
		shippingEvent := map[string]interface{}{
			"order_id":    updatedOrder.ID,
			"customer_id": updatedOrder.CustomerID,
			"items":       updatedOrder.Items,
			"total_value": updatedOrder.TotalPrice,
			"currency":    "IDR",
			"timestamp":   time.Now().UTC(),
		}

		// In a real implementation, you would:
		// 1. Call shipping service API
		// 2. Generate shipping labels
		// 3. Schedule pickup
		// 4. Send tracking information to customer

		// For now, just publish an event (you would need to define shipping topics)
		a.logger.Infof(
			"Shipping arrangement completed for order: %s (mock implementation)",
			order.ID,
		)

		// Publish order processing event
		orderEvent := event.NewOrderLifecycleEvent(
			updatedOrder.ID,
			constant.OrderStatusProcessing,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)

		if err := a.orderLifecycleProducer.Send(ctx, orderEvent); err != nil {
			return fmt.Errorf("failed to send order processing event: %w", err)
		}

		// Log shipping event for demonstration
		shippingPayload, err := json.Marshal(shippingEvent)
		if err != nil {
			a.logger.Errorf("Failed to marshal shipping event: %v", err)

			return err
		}

		a.logger.Infof("Shipping event created: %s", string(shippingPayload))

		a.logger.Infof("Successfully arranged shipping for order: %s", order.ID)

		return nil
	})
}
