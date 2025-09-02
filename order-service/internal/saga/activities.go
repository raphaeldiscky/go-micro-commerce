package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OrderPricing represents the pricing details for an order.
type OrderPricing struct {
	TotalPrice    decimal.Decimal
	TotalDiscount decimal.Decimal
	TotalTax      decimal.Decimal
}

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	// Execution
	ValidateProducts(ctx context.Context, order *entity.Order) error
	ReserveProducts(
		ctx context.Context,
		order *entity.Order,
	) (reservedProducts []entity.Product, err error)
	CalculatePricing(ctx context.Context, order *entity.Order) (pricing OrderPricing, err error)
	ProcessPayment(ctx context.Context, order *entity.Order) (paymentID uuid.UUID, err error)
	ConfirmProductsDeduction(
		ctx context.Context,
		order *entity.Order,
	) error
	CreateShipping(
		ctx context.Context,
		order *entity.Order,
	) (shippingID uuid.UUID, trackingNumber string, err error)
	SendOrderConfirmation(ctx context.Context, order *entity.Order, trackingNumber string) error
	// Compensation
	ReleaseProducts(
		ctx context.Context,
		order *entity.Order,
	) error
	RefundPayment(ctx context.Context, order *entity.Order, paymentID uuid.UUID) error
	RestoreProducts(ctx context.Context, order *entity.Order) error
	CancelShipping(ctx context.Context, shippingID uuid.UUID) error
}

// OrderActivitiesImpl implements the OrderActivities interface.
type OrderActivitiesImpl struct {
	dataStore              repository.DataStore
	productClient          client.ProductClientInterface
	paymentRequestProducer kafka.ProducerInterface
	orderLifecycleProducer kafka.ProducerInterface
	logger                 logger.Logger
}

// NewOrderActivities creates a new OrderActivitiesImpl instance.
func NewOrderActivities(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer kafka.ProducerInterface,
	orderLifecycleProducer kafka.ProducerInterface,
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
func (a *OrderActivitiesImpl) ValidateProducts(_ context.Context, order *entity.Order) error {
	a.logger.Infof("Validating products for order: %s", order.ID)

	// In a real implementation, you would:
	// 1. Check if all products exist in catalog
	// 2. Verify product availability and stock levels
	// 3. Validate product prices and variants

	for i := range order.Items {
		item := &order.Items[i]
		if item.Quantity <= 0 {
			return NewBusinessRuleError(
				ValidateProductsStep,
				fmt.Sprintf("invalid quantity %d for product %s", item.Quantity, item.ProductID),
				nil,
			)
		}
	}

	a.logger.Infof("Product validation completed for order: %s", order.ID)

	return nil
}

// ReserveProducts reserves products for the order items.
func (a *OrderActivitiesImpl) ReserveProducts(
	ctx context.Context,
	order *entity.Order,
) (reservedProducts []entity.Product, err error) {
	a.logger.Infof("Reserving products for order: %s", order.ID)

	if a.productClient == nil {
		return nil, NewNonRetriableError(
			ReserveProductsStep,
			"product service is unavailable",
			nil,
		)
	}

	// Prepare reservation items (avoid copying large structs)
	reservationItems := make([]client.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		reservationItems[i] = client.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Reserve products using product service
	reservedProducts, err = a.productClient.ReserveProducts(
		ctx,
		order.IdempotencyKey,
		reservationItems,
	)
	if err != nil {
		a.logger.Errorf("Failed to reserve stock for order %s: %v", order.ID, err)

		// Categorize error based on type
		if isTemporaryError(err) {
			return nil, NewRetriableError(ReserveProductsStep, "temporary service error", err)
		}

		return nil, NewNonRetriableError(
			ReserveProductsStep,
			"stock reservation failed",
			err,
		)
	}

	a.logger.Infof("Successfully reserved stock for order: %s", order.ID)

	return reservedProducts, nil
}

// CalculatePricing calculates order pricing including discounts and taxes.
func (a *OrderActivitiesImpl) CalculatePricing(
	_ context.Context,
	order *entity.Order,
) (OrderPricing, error) {
	a.logger.Infof("Calculating pricing for order: %s", order.ID)

	// In a real implementation, you would call a pricing service
	// For now, we'll calculate a simple mock pricing
	subtotal := decimal.NewFromInt(0)

	for i := range order.Items {
		// This would normally come from product service
		itemPrice := decimal.NewFromFloat(10.0) // Mock price
		itemTotal := itemPrice.Mul(decimal.NewFromInt(order.Items[i].Quantity))
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

	return OrderPricing{
		TotalPrice:    totalPrice,
		TotalDiscount: discountTotal,
		TotalTax:      taxTotal,
	}, nil
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
		paymentEvent := mq.NewPaymentRequestEvent(
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
			EventType:     event.PaymentRequestedEventType,
			Topic:         event.PaymentRequestTopic,
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
		if isTemporaryError(err) {
			return uuid.Nil, NewRetriableError("ProcessPayment", "temporary database error", err)
		}

		return uuid.Nil, NewNonRetriableError("ProcessPayment", "payment processing failed", err)
	}

	return paymentID, nil
}

// ConfirmProductsDeduction confirms stock deduction after successful payment.
func (a *OrderActivitiesImpl) ConfirmProductsDeduction(
	ctx context.Context,
	order *entity.Order,
) error {
	a.logger.Infof(
		"Confirming inventory deduction for order: %s, reservation: %s",
		order.ID,
	)

	if a.productClient == nil {
		return NewNonRetriableError(
			"ConfirmInventoryDeduction",
			"product service is unavailable",
			nil,
		)
	}

	// In a real implementation, you would:
	// 1. Call product service to confirm the reservation
	// 2. Convert reservation to actual inventory deduction
	// 3. Update product availability counts
	// 4. Release the reservation lock

	for i := range order.Items {
		item := &order.Items[i]
		a.logger.Infof(
			"Confirming deduction of %d units of product %s for order %s",
			item.Quantity,
			item.ProductID,
			order.ID,
		)
	}

	deductionItems := make([]client.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		deductionItems[i] = client.ProductReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	_, err := a.productClient.ConfirmProductsDeduction(ctx, deductionItems)
	if err != nil {
		a.logger.Errorf("Failed to confirm stock deduction for order %s: %v", order.ID, err)

		if isTemporaryError(err) {
			return NewRetriableError("ConfirmStockDeduction", "temporary service error", err)
		}

		return NewNonRetriableError("ConfirmStockDeduction", "stock confirmation failed", err)
	}

	a.logger.Infof("Successfully confirmed stock deduction for order: %s", order.ID)

	return nil
}

// CreateShipping creates shipping arrangement for the order.
func (a *OrderActivitiesImpl) CreateShipping(
	_ context.Context,
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
	_ context.Context,
	order *entity.Order,
	trackingNumber string,
) error {
	a.logger.Infof(
		"Sending order confirmation for order: %s with tracking: %s",
		order.ID,
		trackingNumber,
	)

	// In a real implementation, you would:
	// 1. Send email to customer
	// 2. Send SMS notification
	// 3. Push notification to mobile app
	// 4. Update customer notification preferences

	// For now, just log the confirmation
	a.logger.Infof(
		"Order confirmation sent for order: %s with tracking: %s",
		order.ID,
		trackingNumber,
	)

	return nil
}
