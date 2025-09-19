package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// ReleaseProducts releases reserved products.
func (a *orderActivities) ReleaseProducts(
	ctx context.Context,
	order *entity.Order,
) error {
	a.logger.Infof(
		"Releasing inventory reservation for order: %s",
		order.ID,
	)

	if a.productClient == nil {
		a.logger.Warnf(
			"Product service unavailable, cannot release reservation for order: %s",
			order.ID,
		)

		return nil // Don't fail compensation if service is down
	}

	// Prepare release items
	releaseItems := make([]dto.ProductRestorationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		releaseItems[i] = dto.ProductRestorationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Use the context as-is since it should already contain authentication info
	// The WorkflowContext.AuthenticatedContext() method handles auth injection
	err := a.productClient.ReleaseProducts(ctx, releaseItems)
	if err != nil {
		a.logger.Warnf(
			"Failed to release reservation for order %s: %v (compensation may be incomplete)",
			order.ID,
			err,
		)
		// Still return success to allow compensation to continue
	}

	a.logger.Infof("Successfully released inventory reservation for order: %s", order.ID)

	return nil
}

// RefundPayment refunds payment for the order.
func (a *orderActivities) RefundPayment(
	ctx context.Context,
	order *entity.Order,
	paymentID uuid.UUID,
) error {
	a.logger.Infof("Refunding payment for order: %s, payment ID: %s", order.ID, paymentID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create refund request event
		refundEvent := map[string]any{
			"order_id":    order.ID,
			"customer_id": order.CustomerID,
			"amount":      order.TotalPrice,
			"currency":    order.Currency,
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
			EventType:     kafka.PaymentRefundedEventType,
			Topic:         kafka.PaymentRequestTopic, // Use same topic with different event type
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create refund request event: %w", err)
		}

		a.logger.Infof("Successfully created refund request for order: %s", order.ID)

		return nil
	})
}

// CancelShipping cancels shipping arrangement.
func (a *orderActivities) CancelShipping(_ context.Context, shippingID uuid.UUID) error {
	a.logger.Infof("Canceling shipping: %s", shippingID)

	// In a real implementation, you would call a shipping service API
	// to cancel the shipping arrangement

	// For now, just log the cancellation
	a.logger.Infof("Shipping %s canceled (mock implementation)", shippingID)

	return nil
}

// RestoreProducts restores stock in case of compensation.
func (a *orderActivities) RestoreProducts(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Restoring products for order: %s", order.ID)

	if a.productClient == nil {
		a.logger.Warnf(
			"Product service unavailable, cannot restore stock for order: %s",
			order.ID,
		)

		return nil // Don't fail compensation if service is down
	}

	// Prepare restoration items
	restorationItems := make([]dto.ProductRestorationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		restorationItems[i] = dto.ProductRestorationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Use the context as-is since it should already contain authentication info
	// The WorkflowContext.AuthenticatedContext() method handles auth injection
	_, err := a.productClient.RestoreProducts(ctx, restorationItems)
	if err != nil {
		a.logger.Warnf(
			"Failed to restore stocks for order %s: %v (compensation may be incomplete)",
			order.ID,
			err,
		)
		// Still return success to allow compensation to continue
	}

	a.logger.Infof("Successfully restored stock for order: %s", order.ID)

	return nil
}
