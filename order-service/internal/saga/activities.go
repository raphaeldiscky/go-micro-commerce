package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

type PricingResult struct {
	TotalPrice decimal.Decimal
	Discount   decimal.Decimal
	Tax        decimal.Decimal
}

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	// Execution
	ValidateProducts(ctx context.Context, order *entity.Order) error
	ReserveInventory(ctx context.Context, order *entity.Order) (reservationID uuid.UUID, err error)
	CalculatePricing(ctx context.Context, order *entity.Order) (pricing PricingResult, err error)
	ProcessPayment(ctx context.Context, order *entity.Order) (paymentID uuid.UUID, err error)
	UpdateInventory(ctx context.Context, order *entity.Order) error
	CreateShipping(
		ctx context.Context,
		order *entity.Order,
	) (shippingID uuid.UUID, trackingNumber string, err error)
	SendOrderConfirmation(ctx context.Context, order *entity.Order, trackingNumber string) error
	// Compensation
	ReleaseInventoryReservation(
		ctx context.Context,
		order *entity.Order,
		reservationID uuid.UUID,
	) error
	ConfirmInventoryDeduction(
		ctx context.Context,
		order *entity.Order,
		reservationID uuid.UUID,
	) error
	RestoreInventory(ctx context.Context, order *entity.Order) error
	RefundPayment(ctx context.Context, order *entity.Order, paymentID uuid.UUID) error
	CancelShipping(ctx context.Context, shippingID uuid.UUID) error
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

// ValidateProducts validates product existence and availability.
func (a *OrderActivitiesImpl) ValidateProducts(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Validating products for order: %s", order.ID)

	return nil
}

// CalculatePricing calculates order pricing including discounts and taxes.
func (a *OrderActivitiesImpl) CalculatePricing(
	ctx context.Context,
	order *entity.Order,
) (PricingResult, error) {
	a.logger.Infof("Calculating pricing for order: %s", order.ID)

	// In a real implementation, you would call a pricing service
	// For now, we'll calculate a simple mock pricing
	subtotal := decimal.NewFromInt(0)

	for _, item := range order.Items {
		// This would normally come from product service
		itemPrice := decimal.NewFromFloat(10.0) // Mock price
		itemTotal := itemPrice.Mul(decimal.NewFromInt(item.Quantity))
		subtotal = subtotal.Add(itemTotal)
	}

	// Mock calculations
	taxRate := decimal.NewFromFloat(0.1) // 10% tax
	taxTotal := subtotal.Mul(taxRate)

	// Apply discount if order is large enough
	discountTotal := decimal.NewFromInt(0)
	if subtotal.GreaterThan(decimal.NewFromFloat(100.0)) {
		discountTotal = subtotal.Mul(decimal.NewFromFloat(0.1)) // 10% discount
	}

	totalPrice := subtotal.Add(taxTotal).Sub(discountTotal)

	// Update order with calculated prices
	order.TotalTax = taxTotal
	order.TotalDiscount = discountTotal
	order.TotalPrice = totalPrice

	// Update order items with prices (in real implementation, this would come from product service)
	for i := range order.Items {
		order.Items[i].Price = decimal.NewFromFloat(10.0) // Mock price
	}

	a.logger.Infof("Pricing calculated for order %s: TaxTotal=%s, DiscountTotal=%s, TotalPrice=%s",
		order.ID, subtotal.String(), taxTotal.String(), discountTotal.String(), totalPrice.String())

	return PricingResult{
		TotalPrice: totalPrice,
		Discount:   discountTotal,
		Tax:        taxTotal,
	}, nil
}

// CreateShipping creates shipping arrangement for the order.
func (a *OrderActivitiesImpl) CreateShipping(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, string, error) {
	a.logger.Infof("Creating shipping for order: %s", order.ID)

	// In a real implementation, you would call a shipping service API
	// For now, we'll generate mock shipping data
	shippingID := uuid.New()
	trackingNumber := fmt.Sprintf("TRK-%s-%d", order.ID.String()[:8], time.Now().Unix())

	a.logger.Infof("Successfully created shipping for order %s: ID=%s, Tracking=%s",
		order.ID, shippingID, trackingNumber)

	return shippingID, trackingNumber, nil
}

// SendOrderConfirmation sends order confirmation to customer.
func (a *OrderActivitiesImpl) SendOrderConfirmation(
	ctx context.Context,
	order *entity.Order,
	trackingNumber string,
) error {
	a.logger.Infof("Sending order confirmation for order: %s", order.ID)

	// In a real implementation, you would:
	// 1. Send email to customer
	// 2. Send SMS notification

	// For now, just log the confirmation
	a.logger.Infof("TODO: sent order confirmation for order: %s", order.ID)

	return nil
}

// ReserveInventory reserves inventory for the order items.
func (a *OrderActivitiesImpl) ReserveInventory(
	ctx context.Context,
	order *entity.Order,
) (reservationID uuid.UUID, err error) {
	a.logger.Infof("Reserving inventory for order: %s", order.ID)

	if a.productClient == nil {
		return uuid.Nil, fmt.Errorf("product service is unavailable")
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
	_, err = a.productClient.ReserveProducts(ctx, order.IdempotencyKey, reservationItems)
	if err != nil {
		a.logger.Errorf("Failed to reserve inventory for order %s: %v", order.ID, err)

		return uuid.Nil, fmt.Errorf("inventory reservation failed: %w", err)
	}

	a.logger.Infof("Successfully reserved inventory for order: %s", order.ID)

	return reservationID, nil
}

// ConfirmInventoryDeduction confirms inventory deduction after successful payment.
func (a *OrderActivitiesImpl) ConfirmInventoryDeduction(
	ctx context.Context,
	order *entity.Order,
	reservationID uuid.UUID,
) error {
	a.logger.Infof(
		"Confirming inventory deduction for order: %s, reservation: %s",
		order.ID,
		reservationID,
	)

	if a.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	// Prepare deduction items
	// deductionItems := make([]client.ProductDeductionItem, len(order.Items))
	// for i, item := range order.Items {
	// 	deductionItems[i] = client.ProductDeductionItem{
	// 		ProductID: item.ProductID,
	// 		Quantity:  item.Quantity,
	// 	}
	// }

	// // Confirm inventory deduction using product service
	// err := a.productClient.ConfirmInventoryDeduction(ctx, reservationID, deductionItems)
	// if err != nil {
	// 	a.logger.Errorf("Failed to confirm inventory deduction for order %s: %v", order.ID, err)
	// 	return fmt.Errorf("inventory confirmation failed: %w", err)
	// }

	a.logger.Infof("Successfully confirmed inventory deduction for order: %s", order.ID)

	return nil
}

// ProcessPayment processes payment for the order.
func (a *OrderActivitiesImpl) ProcessPayment(
	ctx context.Context,
	order *entity.Order,
) (paymentID uuid.UUID, err error) {
	a.logger.Infof("Processing payment for order: %s", order.ID)

	// Generate a payment ID upfront
	paymentID = uuid.New()

	err = a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
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

		a.logger.Infof(
			"Successfully created payment request for order: %s with payment ID: %s",
			order.ID,
			paymentID,
		)

		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return paymentID, nil
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
