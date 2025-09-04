// Package temporal provides workflow for create order
package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/activity"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	// Execution
	ValidateProducts(ctx context.Context, order *entity.Order) error
	ReserveProducts(ctx context.Context, order *entity.Order) ([]entity.Product, error)
	CalculatePricing(ctx context.Context, order *entity.Order) (entity.OrderPricing, error)
	ProcessPayment(ctx context.Context, order *entity.Order) (uuid.UUID, error)
	ConfirmProductsDeduction(ctx context.Context, order *entity.Order) error
	CreateShipping(ctx context.Context, order *entity.Order) (dto.CreateShippingResponse, error)
	SendOrderConfirmation(ctx context.Context, req dto.SendOrderConfirmationRequest) error

	// Compensation
	ReleaseProducts(ctx context.Context, order *entity.Order) error
	RefundPayment(ctx context.Context, req dto.RefundPaymentGatewayRequest) error
	RestoreProducts(ctx context.Context, order *entity.Order) error
	CancelShipping(ctx context.Context, shippingID uuid.UUID) error
}

// OrderActivitiesImpl implements order saga activities for Temporal workflows.
type OrderActivitiesImpl struct {
	dataStore     repository.DataStore
	productClient client.ProductClientInterface
}

// NewTemporalActivities creates a new OrderActivities instance.
func NewTemporalActivities(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
) OrderActivities {
	return &OrderActivitiesImpl{
		dataStore:     dataStore,
		productClient: productClient,
	}
}

// ValidateProducts validates products for the order.
func (ta *OrderActivitiesImpl) ValidateProducts(ctx context.Context, order *entity.Order) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ValidateProducts", "orderID", order.ID)

	for i := range order.Items {
		item := &order.Items[i]
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity %d for product %s", item.Quantity, item.ProductID)
		}
	}

	logger.Info("Product validation completed", "orderID", order.ID)

	return nil
}

// ReserveProducts reserves products for the order.
func (ta *OrderActivitiesImpl) ReserveProducts(
	ctx context.Context,
	order *entity.Order,
) ([]entity.Product, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ReserveProducts", "orderID", order.ID)

	if ta.productClient == nil {
		return nil, fmt.Errorf("product service is unavailable")
	}

	// Prepare reservation items
	reservationItems := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		reservationItems[i] = dto.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Reserve products using product service
	reservedProducts, err := ta.productClient.ReserveProducts(
		ctx,
		order.IdempotencyKey,
		reservationItems,
	)
	if err != nil {
		logger.Error("Failed to reserve stock", "orderID", order.ID, "error", err)

		return nil, fmt.Errorf("stock reservation failed: %w", err)
	}

	logger.Info("Successfully reserved stock", "orderID", order.ID)

	return reservedProducts, nil
}

// CalculatePricing calculates pricing for the order.
func (ta *OrderActivitiesImpl) CalculatePricing(
	ctx context.Context,
	order *entity.Order,
) (entity.OrderPricing, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CalculatePricing", "orderID", order.ID)

	// Calculate subtotal
	subtotal := decimal.NewFromInt(0)

	for i := range order.Items {
		// In a real implementation, this would come from product service
		itemPrice := decimal.NewFromFloat(10.0) // Mock price
		itemTotal := itemPrice.Mul(decimal.NewFromInt(order.Items[i].Quantity))
		subtotal = subtotal.Add(itemTotal)
	}

	// Calculate tax and discount
	taxRate := decimal.NewFromFloat(0.1) // 10% tax
	taxTotal := subtotal.Mul(taxRate)

	discountTotal := decimal.NewFromInt(0)
	if subtotal.GreaterThan(decimal.NewFromFloat(100.0)) {
		discountTotal = subtotal.Mul(decimal.NewFromFloat(0.1)) // 10% discount
	}

	totalPrice := subtotal.Add(taxTotal).Sub(discountTotal)

	// Update order with calculated prices
	order.TotalTax = taxTotal
	order.TotalDiscount = discountTotal
	order.TotalPrice = totalPrice

	// Update order items with prices
	for i := range order.Items {
		order.Items[i].UnitPrice = decimal.NewFromFloat(10.0) // Mock price
	}

	logger.Info("Pricing calculated", "orderID", order.ID, "totalPrice", totalPrice.String())

	return entity.OrderPricing{
		TotalPrice:    totalPrice,
		TotalDiscount: discountTotal,
		TotalTax:      taxTotal,
	}, nil
}

// ProcessPayment processes payment for the order using direct database operations.
func (ta *OrderActivitiesImpl) ProcessPayment(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ProcessPayment", "orderID", order.ID)

	// Generate a payment ID
	paymentID := uuid.New()

	// In Temporal, we can directly update the order status instead of using Kafka
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Update order status to processing payment
		order.Status = "processing_payment"

		_, err := orderRepo.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// In a real implementation, you would:
		// 1. Call external payment service
		// 2. Store payment record in database
		// 3. Handle payment response

		logger.Info("Payment processing initiated", "orderID", order.ID, "paymentID", paymentID)

		return nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("payment processing failed: %w", err)
	}

	return paymentID, nil
}

// ConfirmProductsDeduction confirms stock deduction after successful payment.
func (ta *OrderActivitiesImpl) ConfirmProductsDeduction(
	ctx context.Context,
	order *entity.Order,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ConfirmProductsDeduction", "orderID", order.ID)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	deductionItems := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		deductionItems[i] = dto.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	_, err := ta.productClient.ConfirmProductsDeduction(ctx, deductionItems)
	if err != nil {
		logger.Error("Failed to confirm stock deduction", "orderID", order.ID, "error", err)

		return fmt.Errorf("stock confirmation failed: %w", err)
	}

	logger.Info("Successfully confirmed stock deduction", "orderID", order.ID)

	return nil
}

// CreateShipping creates shipping for the order.
func (ta *OrderActivitiesImpl) CreateShipping(
	ctx context.Context,
	order *entity.Order,
) (dto.CreateShippingResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CreateShipping", "orderID", order.ID)

	// In a real implementation, you would call a shipping service API
	shippingID := uuid.New()
	trackingNumber := fmt.Sprintf("TRK-%s-%d", order.ID.String()[:8], time.Now().Unix())

	logger.Info(
		"Successfully created shipping",
		"orderID",
		order.ID,
		"shippingID",
		shippingID,
		"trackingNumber",
		trackingNumber,
	)

	return dto.CreateShippingResponse{
		ShippingID:     shippingID,
		TrackingNumber: trackingNumber,
	}, nil
}

// SendOrderConfirmation sends order confirmation to customer.
func (ta *OrderActivitiesImpl) SendOrderConfirmation(
	ctx context.Context,
	req dto.SendOrderConfirmationRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing SendOrderConfirmation",
		"orderID",
		req.Order.ID,
		"trackingNumber",
		req.TrackingNumber,
	)

	// In a real implementation, you would:
	// 1. Send email to customer
	// 2. Send SMS notification
	// 3. Push notification to mobile app
	// 4. Update notification preferences

	// Update order status to completed
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		req.Order.Status = "completed"
		_, err := orderRepo.Update(ctx, req.Order)

		return err
	})
	if err != nil {
		logger.Error(
			"Failed to update order status to completed",
			"orderID",
			req.Order.ID,
			"error",
			err,
		)

		return fmt.Errorf("failed to complete order: %w", err)
	}

	logger.Info(
		"Order confirmation sent",
		"orderID",
		req.Order.ID,
		"trackingNumber",
		req.TrackingNumber,
	)

	return nil
}
