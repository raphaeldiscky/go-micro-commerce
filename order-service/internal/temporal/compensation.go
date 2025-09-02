package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// ReleaseProducts releases products (compensation for ReserveProducts).
func (ta *OrderActivitiesImpl) ReleaseProducts(
	ctx context.Context,
	order *entity.Order,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ReleaseProducts (compensation)", "orderID", order.ID)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// In a real implementation, you would call product service to release reservations
	logger.Info("Products released successfully", "orderID", order.ID)

	return nil
}

// RefundPayment refunds payment (compensation for ProcessPayment).
func (ta *OrderActivitiesImpl) RefundPayment(
	ctx context.Context,
	req dto.RefundPaymentRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing RefundPayment (compensation)",
		"orderID",
		req.Order.ID,
		"paymentID",
		req.PaymentID,
	)

	// Update order status to refunded
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		req.Order.Status = "refunded"
		_, err := orderRepo.Update(ctx, req.Order)

		return err
	})
	if err != nil {
		logger.Error(
			"Failed to update order status to refunded",
			"orderID",
			req.Order.ID,
			"error",
			err,
		)

		return fmt.Errorf("failed to process refund: %w", err)
	}

	logger.Info(
		"Payment refunded successfully",
		"orderID",
		req.Order.ID,
		"paymentID",
		req.PaymentID,
	)

	return nil
}

// RestoreProducts restores products (compensation for ConfirmProductsDeduction).
func (ta *OrderActivitiesImpl) RestoreProducts(
	ctx context.Context,
	order *entity.Order,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing RestoreProducts (compensation)", "orderID", order.ID)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// In a real implementation, you would call product service to restore inventory
	logger.Info("Products restored successfully", "orderID", order.ID)

	return nil
}

// CancelShipping cancels shipping (compensation for CreateShipping).
func (ta *OrderActivitiesImpl) CancelShipping(
	ctx context.Context,
	shippingID uuid.UUID,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CancelShipping (compensation)", "shippingID", shippingID)

	// In a real implementation, you would call shipping service to cancel shipment
	logger.Info("Shipping canceled successfully", "shippingID", shippingID)

	return nil
}
