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

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	// Execution
	ReserveProductsAndCalculate(
		ctx context.Context,
		order *entity.Order,
	) (calculatedOrder *entity.Order, reservedProducts []entity.Product, err error)
	UpdateOrderPrices(ctx context.Context, order *entity.Order) error
	ProcessPayment(ctx context.Context, order *entity.Order) (paymentID uuid.UUID, err error)
	ConfirmProductsDeduction(
		ctx context.Context,
		order *entity.Order,
		reservedProducts []entity.Product,
	) error
	ProcessFulfillment(
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
	dataStore                  repository.DataStore
	productClient              client.ProductClientInterface
	paymentRequestProducer     kafka.ProducerInterface
	orderLifecycleProducer     kafka.ProducerInterface
	fulfillmentRequestProducer kafka.ProducerInterface
	fulfillmentClient          client.FulfillmentClientInterface
	paymentClient              client.PaymentClientInterface
	logger                     logger.Logger
}

// NewOrderActivities creates a new OrderActivitiesImpl instance.
func NewOrderActivities(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer kafka.ProducerInterface,
	orderLifecycleProducer kafka.ProducerInterface,
	fulfillmentRequestProducer kafka.ProducerInterface,
	fulfillmentClient client.FulfillmentClientInterface,
	paymentClient client.PaymentClientInterface,
	appLogger logger.Logger,
) OrderActivities {
	return &OrderActivitiesImpl{
		dataStore:                  dataStore,
		productClient:              productClient,
		paymentRequestProducer:     paymentRequestProducer,
		orderLifecycleProducer:     orderLifecycleProducer,
		fulfillmentRequestProducer: fulfillmentRequestProducer,
		fulfillmentClient:          fulfillmentClient,
		paymentClient:              paymentClient,
		logger:                     appLogger,
	}
}

// ReserveProductsAndCalculate reserves products for the order items.
func (a *OrderActivitiesImpl) ReserveProductsAndCalculate(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, []entity.Product, error) {
	a.logger.Infof("Get and Reserving products for product: %s")

	if a.productClient == nil {
		return nil, nil, NewNonRetriableError(
			constant.ReserveProductsAndCalculateStep,
			"product service is unavailable",
			nil,
		)
	}

	productIDs := make([]uuid.UUID, len(order.Items))
	for i := range order.Items {
		productIDs[i] = order.Items[i].ProductID
	}

	products, err := a.productClient.GetProducts(ctx, productIDs)
	if err != nil {
		return nil, nil, err
	}

	if len(products) != len(productIDs) {
		return nil, nil, NewNonRetriableError(
			constant.ReserveProductsAndCalculateStep,
			"not all products found",
			nil,
		)
	}

	// Create product map for quick lookup
	productMap := make(map[uuid.UUID]*entity.Product)
	for i, product := range products {
		productMap[product.ID] = &products[i]
	}

	// Prepare reservation items (avoid copying large structs)
	reservations := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		product, exists := productMap[item.ProductID]

		if !exists {
			return nil, nil, NewNonRetriableError(
				constant.ReserveProductsAndCalculateStep,
				fmt.Sprintf("product %s not found", item.ProductID),
				nil,
			)
		}

		reservations[i] = dto.ProductReservationItem{
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			ExpectedVersion: product.Version,
		}
	}

	// Reserve products using product service
	reservedProducts, err := a.productClient.ReserveProducts(
		ctx,
		order.IdempotencyKey,
		reservations,
	)
	if err != nil {
		a.logger.Errorf("Failed to reserve stock for order key %s: %v", order.IdempotencyKey, err)

		// Categorize error based on type
		if isTemporaryError(err) {
			return nil, nil, NewRetriableError(
				constant.ReserveProductsAndCalculateStep,
				"temporary service error",
				err,
			)
		}

		return nil, nil, NewNonRetriableError(
			constant.ReserveProductsAndCalculateStep,
			"stock reservation failed",
			err,
		)
	}

	var orderItems []entity.OrderItem

	for i, product := range reservedProducts {
		orderItem, err := entity.NewOrderItem(
			product.ID,
			order.Items[i].Quantity,
			product.UnitPrice,
		)
		if err != nil {
			return nil, nil, err
		}

		orderItems = append(orderItems, *orderItem)
	}

	// Create domain entity
	calculatedOrder, err := entity.NewOrder(
		order.CustomerID,
		order.IdempotencyKey,
		"IDR",
		orderItems,
	)
	if err != nil {
		return nil, nil, err
	}

	a.logger.Infof("Successfully reserved stock for order entity: %s", calculatedOrder)

	return calculatedOrder, reservedProducts, nil
}

// UpdateOrderPrices updates the order with calculated prices in the database.
func (a *OrderActivitiesImpl) UpdateOrderPrices(ctx context.Context, order *entity.Order) error {
	a.logger.Infof("Updating order prices in database for order: %s", order.ID)

	return a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Update the order with calculated prices
		updatedOrder, err := orderRepo.Update(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to update order prices: %w", err)
		}

		a.logger.Infof(
			"Successfully updated order %s with prices: total=%s, tax=%s, discount=%s",
			updatedOrder.ID,
			updatedOrder.TotalPrice.String(),
			updatedOrder.TotalTax.String(),
			updatedOrder.TotalDiscount.String(),
		)

		return nil
	})
}

// ProcessPayment processes payment for the order.
func (a *OrderActivitiesImpl) ProcessPayment(
	ctx context.Context,
	order *entity.Order,
) (paymentID uuid.UUID, err error) {
	a.logger.Infof("Processing payment for order: %s", order.ID)

	// Step 1: Create and publish payment request event
	err = a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		a.logger.Infof("====> ORDER ID: %s, TOTAL PRICE: %s", order.ID, order.TotalPrice.String())
		// Create payment request event
		paymentEvent := producer.NewPaymentRequestEvent(
			order.ID,
			order.CustomerID,
			order.TotalPrice,
			"IDR",                            // Default currency
			constant.PaymentMethodCreditCard, // Default payment method for saga
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
			EventType:     kafka.PaymentRequestedEventType,
			Topic:         kafka.PaymentRequestTopic,
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
	if err != nil {
		a.logger.Errorf("Failed to publish payment request for order %s: %v", order.ID, err)

		return uuid.Nil, fmt.Errorf("failed to create payment request: %w", err)
	}

	// Step 2: Wait for payment service response
	a.logger.Infof("Waiting for payment response for order: %s", order.ID)

	response, err := a.paymentClient.WaitForPaymentResponse(
		ctx,
		order.ID,
		30*time.Second,
	)
	if err != nil {
		a.logger.Errorf("Failed to receive payment response for order %s: %v", order.ID, err)

		return uuid.Nil, fmt.Errorf("failed to receive payment response: %w", err)
	}

	a.logger.Infof("Successfully received payment response for order %s: ID=%s, Status=%s",
		order.ID, response.PaymentID, response.Status)

	return response.PaymentID, nil
}

// ConfirmProductsDeduction confirms stock deduction after successful payment.
func (a *OrderActivitiesImpl) ConfirmProductsDeduction(
	ctx context.Context,
	order *entity.Order,
	reservedProducts []entity.Product,
) error {
	a.logger.Infof(
		"Confirming inventory deduction for order: %s",
		order.ID,
	)

	if a.productClient == nil {
		return NewNonRetriableError(
			"ConfirmInventoryDeduction",
			"product service is unavailable",
			nil,
		)
	}

	for i := range order.Items {
		item := &order.Items[i]
		a.logger.Infof(
			"Confirming deduction of %d units of product %s for order %s",
			item.Quantity,
			item.ProductID,
			order.ID,
		)
	}

	deductionItems := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		orderItem := &order.Items[i]
		product := &reservedProducts[i]
		deductionItems[i] = dto.ProductReservationItem{
			ProductID:       orderItem.ProductID,
			Quantity:        orderItem.Quantity,
			ExpectedVersion: product.Version,
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

// ProcessFulfillment creates shipping/fulfillment arrangement for the order.
func (a *OrderActivitiesImpl) ProcessFulfillment(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, string, error) {
	a.logger.Infof("Creating shipping for order: %s", order.ID)

	err := a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create mock shipping address (in real implementation, this would come from order data)
		shippingAddress := event.ShippingAddressPayload{
			Street:     "123 Main Street",
			City:       "Jakarta",
			State:      "DKI Jakarta",
			PostalCode: "12345",
			Country:    "Indonesia",
		}

		// Create fulfillment request event
		fulfillmentEvent := producer.NewFulfillmentRequestEvent(order, &shippingAddress)

		payload, err := json.Marshal(fulfillmentEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal fulfillment request event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "fulfillment",
			AggregateID:   order.ID,
			EventType:     kafka.FulfillmentRequestedEventType,
			Topic:         kafka.FulfillmentRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create fulfillment request event: %w", err)
		}

		a.logger.Infof("Successfully created fulfillment request for order: %s", order.ID)

		return nil
	})
	if err != nil {
		a.logger.Errorf("Failed to publish fulfillment request for order %s: %v", order.ID, err)

		return uuid.Nil, "", fmt.Errorf("failed to create shipping: %w", err)
	}

	// Step 2: Wait for fulfillment service response
	a.logger.Infof("Waiting for fulfillment response for order: %s", order.ID)

	response, err := a.fulfillmentClient.WaitForFulfillmentResponse(
		ctx,
		order.ID,
		30*time.Second,
	)
	if err != nil {
		a.logger.Errorf("Failed to receive fulfillment response for order %s: %v", order.ID, err)

		return uuid.Nil, "", fmt.Errorf("failed to receive fulfillment response: %w", err)
	}

	a.logger.Infof("Successfully received fulfillment response for order %s: ID=%s, Tracking=%s",
		order.ID, response.FulfillmentID, response.TrackingNumber)

	return response.FulfillmentID, response.TrackingNumber, nil
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
