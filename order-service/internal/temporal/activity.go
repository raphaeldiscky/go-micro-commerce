// Package temporal provides workflow for create order
package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"go.temporal.io/sdk/activity"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

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
	ReserveProducts(
		ctx context.Context,
		req dto.ReserveProductsRequest,
	) (dto.ReserveProductsResponse, error)
	GetShippingCost(
		ctx context.Context,
		req dto.GetShippingCostRequest,
	) (dto.GetShippingCostResponse, error)
	SetFinalOrderPrices(
		ctx context.Context,
		req dto.SetFinalOrderPricesRequest,
	) (dto.SetFinalOrderPricesResponse, error)
	CreatePayment(ctx context.Context, order *entity.Order) (uuid.UUID, error)
	SendPaymentRequiredNotification(
		ctx context.Context,
		req dto.SendPaymentRequiredNotificationRequest,
	) error
	SendPaymentReminderNotification(
		ctx context.Context,
		req dto.SendPaymentReminderNotificationRequest,
	) error
	WaitForPaymentConfirmation(
		ctx context.Context,
		req dto.WaitForPaymentConfirmationRequest,
	) (dto.WaitForPaymentConfirmationResponse, error)
	ProcessFulfillment(
		ctx context.Context,
		order *entity.Order,
		shipping *dto.Shipping,
	) (dto.ProcessFulfillmentResponse, error)
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

// orderActivities implements order saga activities for Temporal workflows.
type orderActivities struct {
	dataStore         repository.DataStore
	productClient     client.ProductClient
	fulfillmentClient client.FulfillmentClient
	paymentClient     client.PaymentClient
}

// NewTemporalActivities creates a new OrderActivities instance.
func NewTemporalActivities(
	dataStore repository.DataStore,
	productClient client.ProductClient,
	fulfillmentClient client.FulfillmentClient,
	paymentClient client.PaymentClient,
) OrderActivities {
	return &orderActivities{
		dataStore:         dataStore,
		productClient:     productClient,
		fulfillmentClient: fulfillmentClient,
		paymentClient:     paymentClient,
	}
}

// ReserveProducts reserves products for the order items and calculates order details.
func (ta *orderActivities) ReserveProducts(
	ctx context.Context,
	req dto.ReserveProductsRequest,
) (dto.ReserveProductsResponse, error) {
	logger := activity.GetLogger(ctx)
	order := req.Order
	userAuth := req.UserAuth

	logger.Info("Executing ReserveProductsAndCalculate", "orderID", order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, userAuth)

	if ta.productClient == nil {
		return dto.ReserveProductsResponse{}, errors.New("product service is unavailable")
	}

	productIDs := make([]uuid.UUID, len(order.Items))
	for i := range order.Items {
		productIDs[i] = order.Items[i].ProductID
	}

	products, err := ta.productClient.GetProducts(ctx, productIDs)
	if err != nil {
		logger.Error(
			"Failed to get products from product service",
			"productIDs",
			productIDs,
			"error",
			err,
		)

		return dto.ReserveProductsResponse{}, err
	}

	if len(products) != len(productIDs) {
		return dto.ReserveProductsResponse{}, fmt.Errorf(
			"not all products found: requested %d, found %d",
			len(productIDs),
			len(products),
		)
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
			return dto.ReserveProductsResponse{}, fmt.Errorf(
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

		return dto.ReserveProductsResponse{}, fmt.Errorf(
			"stock reservation failed: %w",
			err,
		)
	}

	var orderItems []entity.OrderItem

	for i, product := range reservedProducts {
		orderItem, rowErr := entity.NewOrderItem(
			product.ID,
			order.Items[i].Quantity,
			product.UnitPrice,
		)
		if rowErr != nil {
			return dto.ReserveProductsResponse{}, rowErr
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
		return dto.ReserveProductsResponse{}, err
	}

	// Get customer email from user auth
	email := userAuth.Email
	if email == "" {
		return dto.ReserveProductsResponse{}, errors.New("customer email not found in user auth")
	}

	logger.Info("Successfully reserved stock for order", "orderID", order.ID)

	return dto.ReserveProductsResponse{
		CalculatedOrder:  calculatedOrder,
		ReservedProducts: reservedProducts,
		CustomerEmail:    email,
	}, nil
}

// GetShippingCost calculates shipping cost by calling fulfillment service without creating actual shipment.
func (ta *orderActivities) GetShippingCost(
	ctx context.Context,
	req dto.GetShippingCostRequest,
) (dto.GetShippingCostResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing GetShippingCost", "orderID", req.Order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, *req.UserAuth)

	shippingCost, err := ta.fulfillmentClient.GetShippingCost(ctx, req.Order, req.Shipping)
	if err != nil {
		logger.Error("Failed to get shipping cost", "orderID", req.Order.ID, "error", err)

		return dto.GetShippingCostResponse{}, fmt.Errorf("failed to get shipping cost: %w", err)
	}

	logger.Info(
		"Successfully received shipping cost",
		"orderID", req.Order.ID,
	)

	return dto.GetShippingCostResponse{
		ShippingCost: shippingCost,
	}, nil
}

// CreatePayment create payment for the order.
func (ta *orderActivities) CreatePayment(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ProcessPayment", "orderID", order.ID)

	// Step 1: Create and publish payment request event
	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
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
		constant.CreatePaymentStepTimeout,
	)
	if err != nil {
		logger.Error("Failed to receive payment response", "orderID", order.ID, "error", err)

		return uuid.Nil, fmt.Errorf("failed to receive payment response: %w", err)
	}

	logger.Info(
		"Successfully received payment response",
		"orderID",
		order.ID,
	)

	return response.PaymentID, nil
}

// SendPaymentRequiredNotification sends a payment required notification to the customer.
func (ta *orderActivities) SendPaymentRequiredNotification(
	ctx context.Context,
	req dto.SendPaymentRequiredNotificationRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing SendPaymentRequiredNotification", "orderID", req.Order.ID)

	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		notificationEvent := producer.NewNotificationRequestEvent(
			req.Order,
			req.ReservedProducts,
			req.CustomerEmail,
			"Customer Name",
			nil,
			pkgconstant.TemplateOrderPaymentRequired,
			"Payment Required - Complete Your Order",
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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			logger.Error(
				"Failed to create payment required notification",
				"orderID", req.Order.ID,
				"error", err,
			)

			return fmt.Errorf("failed to create payment required notification event: %w", err)
		}

		logger.Info("Successfully created payment required notification", "orderID", req.Order.ID)

		return nil
	})
	if err != nil {
		logger.Error(
			"Failed to create payment required notification",
			"orderID", req.Order.ID,
			"error", err,
		)

		return fmt.Errorf("failed to create payment required notification: %w", err)
	}

	return nil
}

// SendPaymentReminderNotification sends a payment reminder notification to the customer.
func (ta *orderActivities) SendPaymentReminderNotification(
	ctx context.Context,
	req dto.SendPaymentReminderNotificationRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing SendPaymentReminderNotification",
		"orderID", req.Order.ID,
		"sequence", req.ReminderSequence,
	)

	err := ta.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		templateType := pkgconstant.TemplateOrderPaymentReminder

		notificationEvent := producer.NewNotificationRequestEvent(
			req.Order,
			req.ReservedProducts,
			req.CustomerEmail,
			"Customer Name",
			nil,
			templateType,
			req.Subject,
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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create payment reminder notification event: %w", err)
		}

		logger.Info("Successfully created payment reminder notification",
			"orderID", req.Order.ID,
			"sequence", req.ReminderSequence,
		)

		return nil
	})
	if err != nil {
		logger.Error(
			"Failed to create payment reminder notification",
			"orderID", req.Order.ID,
			"sequence", req.ReminderSequence,
			"error", err,
		)

		return fmt.Errorf("failed to create payment reminder notification: %w", err)
	}

	return nil
}

// WaitForPaymentConfirmation waits for payment confirmation with timeout and manages payment reminders using Temporal schedules.
func (ta *orderActivities) WaitForPaymentConfirmation(
	ctx context.Context,
	req dto.WaitForPaymentConfirmationRequest,
) (dto.WaitForPaymentConfirmationResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing WaitForPaymentConfirmation", "orderID", req.Order.ID)

	// Wait for payment confirmation
	response, err := ta.paymentClient.WaitForPaymentResponse(
		ctx,
		req.Order.ID,
		constant.WaitForPaymentConfirmationActivityTimeout,
	)
	if err != nil {
		logger.Error(
			"Failed to receive payment confirmation",
			"orderID",
			req.Order.ID,
			"error",
			err,
		)

		return dto.WaitForPaymentConfirmationResponse{}, fmt.Errorf(
			"failed to receive payment confirmation: %w",
			err,
		)
	}

	logger.Info(
		"Successfully received payment confirmation",
		"orderID", req.Order.ID,
		"paymentID", response.PaymentID,
	)

	return dto.WaitForPaymentConfirmationResponse{
		PaymentID: response.PaymentID,
	}, nil
}

// ProcessFulfillment creates shipping/fulfillment arrangement for the order.
func (ta *orderActivities) ProcessFulfillment(
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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
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
		constant.ProcessFulfillmentStepTimeout,
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
	)

	return dto.ProcessFulfillmentResponse{
		ShippingID:     response.FulfillmentID,
		ShippingCost:   response.ShippingCost,
		TrackingNumber: response.TrackingNumber,
	}, nil
}

// SetFinalOrderPrices updates the order with shipping cost and final prices in the database.
func (ta *orderActivities) SetFinalOrderPrices(
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
func (ta *orderActivities) ConfirmProductsDeduction(
	ctx context.Context,
	req *dto.ConfirmProductsDeductionRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ConfirmProductsDeduction", "orderID", req.Order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return errors.New("product service is unavailable")
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

	deductionItems := make([]dto.ProductRestorationItem, len(req.Order.Items))

	for i := range req.Order.Items {
		orderItem := &req.Order.Items[i]
		deductionItems[i] = dto.ProductRestorationItem{
			ProductID: orderItem.ProductID,
			Quantity:  orderItem.Quantity,
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
func (ta *orderActivities) SendOrderConfirmedNotification(
	ctx context.Context,
	req dto.SendOrderConfirmedNotificationRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing SendOrderConfirmedNotification",
		"orderID", req.Order.ID,
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
			pkgconstant.TemplateOrderConfirmed,
			"Order Confirmed - Payment Received",
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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
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

// ReleaseProducts releases reserved products during compensation.
func (ta *orderActivities) ReleaseProducts(
	ctx context.Context,
	req dto.ReleaseProductsRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing ReleaseProducts compensation", "orderID", req.Order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return errors.New("product service is unavailable")
	}

	releaseItems := make([]dto.ProductRestorationItem, len(req.Order.Items))

	for i := range req.Order.Items {
		orderItem := &req.Order.Items[i]
		releaseItems[i] = dto.ProductRestorationItem{
			ProductID: orderItem.ProductID,
			Quantity:  orderItem.Quantity,
		}
	}

	err := ta.productClient.ReleaseProducts(ctx, releaseItems)
	if err != nil {
		logger.Error("Failed to release reserved products", "orderID", req.Order.ID, "error", err)

		return fmt.Errorf("failed to release reserved products: %w", err)
	}

	logger.Info("Successfully released reserved products", "orderID", req.Order.ID)

	return nil
}

// RefundPayment refunds payment during compensation.
func (ta *orderActivities) RefundPayment(
	ctx context.Context,
	req dto.RefundPaymentGatewayRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"Executing RefundPayment compensation",
		"orderID",
		req.Order.ID,
	)

	// For now, just log the compensation - payment client doesn't have RefundPayment method
	// TODO: Implement actual payment refund when PaymentClient supports it
	logger.Info("Payment refund compensation logged",
		"orderID", req.Order.ID,
		"paymentID", req.PaymentID,
	)

	return nil
}

// RestoreProducts restores deducted products during compensation.
func (ta *orderActivities) RestoreProducts(
	ctx context.Context,
	req dto.RestoreProductsRequest,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing RestoreProducts compensation", "orderID", req.Order.ID)

	// Add user authentication info to context for gRPC calls
	ctx = echoutils.AddUserAuthToContexts(ctx, req.UserAuth)

	if ta.productClient == nil {
		return errors.New("product service is unavailable")
	}

	restoreItems := make([]dto.ProductRestorationItem, len(req.Order.Items))

	for i := range req.Order.Items {
		orderItem := &req.Order.Items[i]
		restoreItems[i] = dto.ProductRestorationItem{
			ProductID: orderItem.ProductID,
			Quantity:  orderItem.Quantity,
		}
	}

	_, err := ta.productClient.RestoreProducts(ctx, restoreItems)
	if err != nil {
		logger.Error("Failed to restore deducted products", "orderID", req.Order.ID, "error", err)

		return fmt.Errorf("failed to restore deducted products: %w", err)
	}

	logger.Info("Successfully restored deducted products", "orderID", req.Order.ID)

	return nil
}

// CancelShipping cancels shipping during compensation.
func (ta *orderActivities) CancelShipping(
	ctx context.Context,
	shippingID uuid.UUID,
) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing CancelShipping compensation", "shippingID", shippingID)

	// For now, just log the compensation - fulfillment client doesn't have CancelShipping method
	// TODO: Implement actual shipping cancellation when FulfillmentClient supports it
	logger.Info("Shipping cancellation compensation logged",
		"shippingID", shippingID,
	)

	return nil
}
