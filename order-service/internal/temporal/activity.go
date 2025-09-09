// Package temporal provides workflow for create order
package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"go.temporal.io/sdk/activity"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// OrderActivities defines the interface for order saga activities to match saga implementation.
type OrderActivities interface {
	// Execution
	ReserveProductsAndCalculate(
		ctx context.Context,
		req dto.ReserveProductsAndCalculateRequest,
	) (dto.ReserveProductsAndCalculateResponse, error)
	ProcessFulfillment(
		ctx context.Context,
		order *entity.Order,
		shipping *dto.Shipping,
	) (dto.ProcessFulfillmentResponse, error)
	SetFinalOrderPrices(
		ctx context.Context,
		req dto.SetFinalOrderPricesRequest,
	) (dto.SetFinalOrderPricesResponse, error)
	ProcessPayment(ctx context.Context, order *entity.Order) (uuid.UUID, error)
	ConfirmProductsDeduction(ctx context.Context, req *dto.ConfirmProductsDeductionRequest) error
	SendOrderConfirmedNotification(
		ctx context.Context,
		req dto.SendOrderConfirmedNotificationRequest,
	) error

	// Compensation
	ReleaseProducts(ctx context.Context, req dto.ReleaseProductsRequest) error
	RefundPayment(ctx context.Context, req dto.RefundPaymentGatewayRequest) error
	RestoreProducts(ctx context.Context, req dto.RestoreProductsRequest) error
	CancelShipping(ctx context.Context, shippingID uuid.UUID) error
}

// OrderActivitiesImpl implements order saga activities for Temporal workflows.
type OrderActivitiesImpl struct {
	dataStore                  repository.DataStore
	productClient              client.ProductClientInterface
	paymentRequestProducer     kafka.ProducerInterface
	orderLifecycleProducer     kafka.ProducerInterface
	fulfillmentRequestProducer kafka.ProducerInterface
	fulfillmentClient          client.FulfillmentClientInterface
	paymentClient              client.PaymentClientInterface
}

// NewTemporalActivities creates a new OrderActivities instance.
func NewTemporalActivities(
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	paymentRequestProducer kafka.ProducerInterface,
	orderLifecycleProducer kafka.ProducerInterface,
	fulfillmentRequestProducer kafka.ProducerInterface,
	fulfillmentClient client.FulfillmentClientInterface,
	paymentClient client.PaymentClientInterface,
) OrderActivities {
	return &OrderActivitiesImpl{
		dataStore:                  dataStore,
		productClient:              productClient,
		paymentRequestProducer:     paymentRequestProducer,
		orderLifecycleProducer:     orderLifecycleProducer,
		fulfillmentRequestProducer: fulfillmentRequestProducer,
		fulfillmentClient:          fulfillmentClient,
		paymentClient:              paymentClient,
	}
}

// ReserveProductsAndCalculate reserves products for the order items and calculates order details.
func (ta *OrderActivitiesImpl) ReserveProductsAndCalculate(
	ctx context.Context,
	req dto.ReserveProductsAndCalculateRequest,
) (dto.ReserveProductsAndCalculateResponse, error) {
	logger := activity.GetLogger(ctx)
	order := req.Order
	userAuth := req.UserAuth

	logger.Info("Executing ReserveProductsAndCalculate", "orderID", order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, userAuth)

	if ta.productClient == nil {
		return dto.ReserveProductsAndCalculateResponse{}, fmt.Errorf(
			"product service is unavailable",
		)
	}

	productIDs := make([]uuid.UUID, len(order.Items))
	for i := range order.Items {
		productIDs[i] = order.Items[i].ProductID
	}

	products, err := ta.productClient.GetProducts(ctx, productIDs)
	if err != nil {
		return dto.ReserveProductsAndCalculateResponse{}, err
	}

	if len(products) != len(productIDs) {
		return dto.ReserveProductsAndCalculateResponse{}, fmt.Errorf("not all products found")
	}

	// Create product map for quick lookup
	productMap := make(map[uuid.UUID]*entity.Product)
	for i, product := range products {
		productMap[product.ID] = &products[i]
	}

	// Prepare reservation items
	reservations := make([]dto.ProductReservationItem, len(order.Items))

	for i := range order.Items {
		item := &order.Items[i]
		product, exists := productMap[item.ProductID]

		if !exists {
			return dto.ReserveProductsAndCalculateResponse{}, fmt.Errorf(
				"product %s not found",
				item.ProductID,
			)
		}

		reservations[i] = dto.ProductReservationItem{
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			ExpectedVersion: product.Version,
		}
	}

	// Reserve products using product service
	reservedProducts, err := ta.productClient.ReserveProducts(
		ctx,
		order.IdempotencyKey,
		reservations,
	)
	if err != nil {
		logger.Error("Failed to reserve stock for order", "orderID", order.ID, "error", err)

		return dto.ReserveProductsAndCalculateResponse{}, fmt.Errorf(
			"stock reservation failed: %w",
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
			return dto.ReserveProductsAndCalculateResponse{}, err
		}

		orderItems = append(orderItems, *orderItem)
	}

	// Create calculated order entity
	calculatedOrder, err := entity.NewOrder(
		order.CustomerID,
		order.IdempotencyKey,
		"IDR",
		orderItems,
	)
	if err != nil {
		return dto.ReserveProductsAndCalculateResponse{}, err
	}

	// Get customer email from user auth
	email := userAuth.Email
	if email == "" {
		return dto.ReserveProductsAndCalculateResponse{}, fmt.Errorf(
			"customer email not found in user auth",
		)
	}

	logger.Info("Successfully reserved stock for order", "orderID", order.ID)

	return dto.ReserveProductsAndCalculateResponse{
		CalculatedOrder:  calculatedOrder,
		ReservedProducts: reservedProducts,
		CustomerEmail:    email,
	}, nil
}

// ProcessPayment processes payment for the order.
func (ta *OrderActivitiesImpl) ProcessPayment(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ProcessPayment", "orderID", order.ID)

	// Step 1: Create and publish payment request event
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		logger.Info(
			"Creating payment request",
			"orderID",
			order.ID,
			"totalPrice",
			order.TotalPrice.String(),
		)

		// Create payment request event
		paymentEvent := producer.NewPaymentRequestEvent(
			order.ID,
			order.CustomerID,
			order.TotalPrice,
			"IDR",
			constant.PaymentMethodCreditCard,
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

		logger.Info("Successfully created payment request for order", "orderID", order.ID)

		return nil
	})
	if err != nil {
		logger.Error("Failed to publish payment request", "orderID", order.ID, "error", err)

		return uuid.Nil, fmt.Errorf("failed to create payment request: %w", err)
	}

	// Step 2: Wait for payment service response
	logger.Info("Waiting for payment response", "orderID", order.ID)

	response, err := ta.paymentClient.WaitForPaymentResponse(
		ctx,
		order.ID,
		30*time.Second,
	)
	if err != nil {
		logger.Error("Failed to receive payment response", "orderID", order.ID, "error", err)

		return uuid.Nil, fmt.Errorf("failed to receive payment response: %w", err)
	}

	logger.Info(
		"Successfully received payment response",
		"orderID",
		order.ID,
		"paymentID",
		response.PaymentID,
		"status",
		response.Status,
	)

	return response.PaymentID, nil
}

// ProcessFulfillment creates shipping/fulfillment arrangement for the order.
func (ta *OrderActivitiesImpl) ProcessFulfillment(
	ctx context.Context,
	order *entity.Order,
	shipping *dto.Shipping,
) (dto.ProcessFulfillmentResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ProcessFulfillment", "orderID", order.ID)

	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create fulfillment request event
		fulfillmentEvent := producer.NewFulfillmentRequestEvent(order, shipping)

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

		logger.Info("Successfully created fulfillment request", "orderID", order.ID)

		return nil
	})
	if err != nil {
		logger.Error("Failed to publish fulfillment request", "orderID", order.ID, "error", err)

		return dto.ProcessFulfillmentResponse{}, fmt.Errorf(
			"failed to create fulfillment request: %w",
			err,
		)
	}

	// Step 2: Wait for fulfillment service response
	logger.Info("Waiting for fulfillment response", "orderID", order.ID)

	response, err := ta.fulfillmentClient.WaitForFulfillmentResponse(
		ctx,
		order.ID,
		30*time.Second,
	)
	if err != nil {
		logger.Error("Failed to receive fulfillment response", "orderID", order.ID, "error", err)

		return dto.ProcessFulfillmentResponse{}, fmt.Errorf(
			"failed to receive fulfillment response: %w",
			err,
		)
	}

	logger.Info(
		"Successfully received fulfillment response",
		"orderID", order.ID,
		"fulfillmentID", response.FulfillmentID,
		"shippingCost", response.ShippingCost,
		"trackingNumber", response.TrackingNumber,
	)

	return dto.ProcessFulfillmentResponse{
		ShippingID:     response.FulfillmentID,
		ShippingCost:   response.ShippingCost,
		TrackingNumber: response.TrackingNumber,
	}, nil
}

// SetFinalOrderPrices updates the order with shipping cost and final prices in the database.
func (ta *OrderActivitiesImpl) SetFinalOrderPrices(
	ctx context.Context,
	req dto.SetFinalOrderPricesRequest,
) (dto.SetFinalOrderPricesResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing SetFinalOrderPrices", "orderID", req.Order.ID)

	var response dto.SetFinalOrderPricesResponse

	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Update the original order with shipping cost
		err := req.Order.UpdateShippingCost(req.ShippingCost)
		if err != nil {
			return err
		}

		// Update the order with calculated prices
		updatedOrder, err := orderRepo.Update(ctx, req.Order)
		if err != nil {
			return fmt.Errorf("failed to update order prices: %w", err)
		}

		logger.Info(
			"Successfully updated order with prices",
			"orderID", updatedOrder.ID,
			"totalPrice", updatedOrder.TotalPrice.String(),
			"totalTax", updatedOrder.TotalTax.String(),
			"totalDiscount", updatedOrder.TotalDiscount.String(),
		)

		// Set the response with the updated order
		response.UpdatedOrder = updatedOrder

		return nil
	})
	if err != nil {
		return dto.SetFinalOrderPricesResponse{}, err
	}

	return response, nil
}

// ConfirmProductsDeduction confirms stock deduction after successful payment.
func (ta *OrderActivitiesImpl) ConfirmProductsDeduction(
	ctx context.Context,
	req *dto.ConfirmProductsDeductionRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ConfirmProductsDeduction", "orderID", req.Order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return fmt.Errorf("product service is unavailable")
	}

	for i := range req.Order.Items {
		item := &req.Order.Items[i]
		logger.Info(
			"Confirming deduction of product",
			"quantity", item.Quantity,
			"productID", item.ProductID,
			"orderID", req.Order.ID,
		)
	}

	deductionItems := make([]dto.ProductReservationItem, len(req.Order.Items))

	for i := range req.Order.Items {
		orderItem := &req.Order.Items[i]
		product := &req.ReservedProducts[i]
		deductionItems[i] = dto.ProductReservationItem{
			ProductID:       orderItem.ProductID,
			Quantity:        orderItem.Quantity,
			ExpectedVersion: product.Version,
		}
	}

	_, err := ta.productClient.ConfirmProductsDeduction(ctx, deductionItems)
	if err != nil {
		logger.Error("Failed to confirm stock deduction", "orderID", req.Order.ID, "error", err)

		return fmt.Errorf("stock confirmation failed: %w", err)
	}

	logger.Info("Successfully confirmed stock deduction", "orderID", req.Order.ID)

	return nil
}

// SendOrderConfirmedNotification sends order confirmation to customer.
func (ta *OrderActivitiesImpl) SendOrderConfirmedNotification(
	ctx context.Context,
	req dto.SendOrderConfirmedNotificationRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing SendOrderConfirmedNotification",
		"orderID", req.Order.ID,
		"trackingNumber", req.TrackingNumber,
		"customerEmail", req.CustomerEmail,
	)

	// Create notification request event
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create order confirmation notification event
		notificationEvent := producer.NewNotificationRequestEvent(
			req.Order,
			req.Products,
			req.CustomerEmail,
			"Customer Name", // TODO: Get actual customer name from user service if needed
			&req.TrackingNumber,
		)

		payload, err := json.Marshal(notificationEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal notification event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "notification",
			AggregateID:   req.Order.ID,
			EventType:     kafka.NotificationRequestedEventType,
			Topic:         kafka.NotificationRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create notification event: %w", err)
		}

		logger.Info("Successfully created order confirmation notification", "orderID", req.Order.ID)

		return nil
	})
	if err != nil {
		logger.Error("Failed to create notification request", "orderID", req.Order.ID, "error", err)

		return fmt.Errorf("failed to send order confirmation: %w", err)
	}

	logger.Info(
		"Order confirmation notification queued",
		"orderID", req.Order.ID,
		"trackingNumber", req.TrackingNumber,
	)

	return nil
}
