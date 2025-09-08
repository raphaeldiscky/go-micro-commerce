package temporal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"go.temporal.io/sdk/activity"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// Compensation Activities

// ReleaseProducts releases reserved products.
func (ta *OrderActivitiesImpl) ReleaseProducts(
	ctx context.Context,
	req dto.ReleaseProductsRequest,
) error {
	logger := activity.GetLogger(ctx)
	order := req.Order
	logger.Info("Executing ReleaseProducts compensation", "orderID", order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// Create reservation items for release
	releaseItems := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		releaseItems[i] = dto.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Release products using product service
	if err := ta.productClient.ReleaseProducts(ctx, releaseItems); err != nil {
		logger.Error("Failed to release products", "orderID", order.ID, "error", err)

		return fmt.Errorf("product release failed: %w", err)
	}

	logger.Info("Successfully released products", "orderID", order.ID)

	return nil
}

// RefundPayment refunds payment for the order.
func (ta *OrderActivitiesImpl) RefundPayment(
	ctx context.Context,
	req dto.RefundPaymentGatewayRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing RefundPayment compensation",
		"orderID",
		req.Order.ID,
		"paymentID",
		req.PaymentID,
	)

	// For Temporal, handle refund by updating order status directly
	// In real implementation, you would call payment service to process refund
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
		"Successfully refunded payment",
		"orderID",
		req.Order.ID,
		"paymentID",
		req.PaymentID,
	)

	return nil
}

// RestoreProducts restores deducted products.
func (ta *OrderActivitiesImpl) RestoreProducts(
	ctx context.Context,
	req dto.RestoreProductsRequest,
) error {
	logger := activity.GetLogger(ctx)
	order := req.Order
	logger.Info("Executing RestoreProducts compensation", "orderID", order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// Create restoration items
	restorationItems := make([]dto.ProductRestorationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		restorationItems[i] = dto.ProductRestorationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Restore products using product service
	_, err := ta.productClient.RestoreProducts(ctx, restorationItems)
	if err != nil {
		logger.Error("Failed to restore products", "orderID", order.ID, "error", err)

		return fmt.Errorf("product restore failed: %w", err)
	}

	logger.Info("Successfully restored products", "orderID", order.ID)

	return nil
}

// CancelShipping cancels shipping for the order.
func (ta *OrderActivitiesImpl) CancelShipping(
	ctx context.Context,
	shippingID uuid.UUID,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CancelShipping compensation", "shippingID", shippingID)

	// For Temporal, we simulate shipping cancellation
	// In real implementation, you would call fulfillment service to cancel shipping
	logger.Info("Shipping canceled successfully", "shippingID", shippingID)

	return nil
}
