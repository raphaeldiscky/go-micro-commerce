package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/task"
)

// OrderActivities defines the interface for order saga activities.
type OrderActivities interface {
	// Execution
	ReserveProductsAndCalculate(
		ctx context.Context,
		order *entity.Order,
	) (calculatedOrder *entity.Order, reservedProducts []entity.Product, err error)
	GetShippingCost(
		ctx context.Context,
		order *entity.Order,
		shipping *dto.Shipping,
	) (shippingCost decimal.Decimal, err error)
	SetFinalOrderPrices(ctx context.Context, order *entity.Order) error
	CreatePayment(ctx context.Context, order *entity.Order) (paymentID uuid.UUID, err error)
	SendPaymentRequiredNotification(
		ctx context.Context,
		order *entity.Order,
		reservedProducts []entity.Product,
		customerEmail string,
	) error
	WaitForPaymentConfirmation(
		ctx context.Context,
		order *entity.Order,
		customerEmail string,
	) (paymentID uuid.UUID, err error) // 24h timeout
	ProcessFulfillment(
		ctx context.Context,
		payload *Payload,
	) (fulfillmentID uuid.UUID, shippingCost decimal.Decimal, trackingNumber string, err error)
	ConfirmProductsDeduction(
		ctx context.Context,
		order *entity.Order,
		reservedProducts []entity.Product,
	) error
	SendOrderConfirmedNotification(
		ctx context.Context,
		order *entity.Order,
		products []entity.Product,
		trackingNumber *string,
		customerEmail string,
	) error
	// Compensation
	ReleaseProducts(
		ctx context.Context,
		order *entity.Order,
	) error
	RefundPayment(ctx context.Context, order *entity.Order, paymentID uuid.UUID) error
	RestoreProducts(ctx context.Context, order *entity.Order) error
	CancelShipping(ctx context.Context, shippingID uuid.UUID) error
}

// orderActivities implements the OrderActivities interface.
type orderActivities struct {
	dataStore              repository.DataStore
	productClient          client.ProductClient
	fulfillmentClient      client.FulfillmentClient
	paymentClient          client.PaymentClient
	asynqClient            asynq.Client
	taskCancellationHelper *task.CancellationHelper
	logger                 logger.Logger
}

// NewOrderActivities creates a new orderActivities instance.
func NewOrderActivities(
	dataStore repository.DataStore,
	productClient client.ProductClient,
	fulfillmentClient client.FulfillmentClient,
	paymentClient client.PaymentClient,
	asynqClient asynq.Client,
	taskCancellationService asynq.TaskCancellationService,
	appLogger logger.Logger,
) OrderActivities {
	taskCancellationHelper := task.NewCancellationHelper(taskCancellationService, appLogger)

	return &orderActivities{
		dataStore:              dataStore,
		productClient:          productClient,
		fulfillmentClient:      fulfillmentClient,
		paymentClient:          paymentClient,
		asynqClient:            asynqClient,
		taskCancellationHelper: taskCancellationHelper,
		logger:                 appLogger,
	}
}

// ReserveProductsAndCalculate reserves products for the order items.
func (a *orderActivities) ReserveProductsAndCalculate(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, []entity.Product, error) {
	a.logger.Infof("Get and Reserving products for order: %s", order.ID)

	productIDs := make([]uuid.UUID, len(order.Items))
	for i := range order.Items {
		productIDs[i] = order.Items[i].ProductID
	}

	// Add user authentication info to context for gRPC calls
	ctx = addUserAuthToContext(ctx, a.logger, order.ID)

	products, err := a.productClient.GetProducts(ctx, productIDs)
	if err != nil {
		return nil, nil, err
	}

	if len(products) != len(productIDs) {
		return nil, nil, NewNonRetriableError(
			constant.ReserveProductsStep,
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
				constant.ReserveProductsStep,
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
		sagaErr := CategorizeError(constant.ReserveProductsStep, err)

		return nil, nil, sagaErr
	}

	orderItems := make([]entity.OrderItem, 0, len(reservedProducts))

	for i, product := range reservedProducts {
		orderItem, rowErr := entity.NewOrderItem(
			product.ID,
			order.Items[i].Quantity,
			product.UnitPrice,
		)
		if rowErr != nil {
			return nil, nil, rowErr
		}

		orderItems = append(orderItems, *orderItem)
	}

	// Create domain entity
	newOrder, err := entity.NewOrder(
		order.CustomerID,
		order.IdempotencyKey,
		order.PaymentGateway,
		order.Currency,
		orderItems,
	)
	if err != nil {
		return nil, nil, err
	}

	a.logger.Infof("Successfully reserved stock for order entity: %s", newOrder)

	return newOrder, reservedProducts, nil
}

// GetShippingCost calculates shipping cost from fulfillment service without creating actual shipment.
func (a *orderActivities) GetShippingCost(
	ctx context.Context,
	order *entity.Order,
	shipping *dto.Shipping,
) (decimal.Decimal, error) {
	a.logger.Infof(
		"Getting shipping cost from fulfillment service for order: %s with shipping details",
		order.ID,
	)

	// Add user authentication info to context for gRPC calls
	ctx = addUserAuthToContext(ctx, a.logger, order.ID)

	shippingCost, err := a.fulfillmentClient.GetShippingCost(ctx, order, shipping)
	if err != nil {
		a.logger.Errorf("Failed to get shipping cost for order %s: %v", order.ID, err)
		return decimal.Zero, fmt.Errorf("failed to get shipping cost: %w", err)
	}

	a.logger.Infof("Successfully received shipping cost for order %s: %s %s",
		order.ID, shippingCost, order.Currency)

	return shippingCost, nil
}

// SetFinalOrderPrices updates the order with shipping cost and final prices in the database.
func (a *orderActivities) SetFinalOrderPrices(ctx context.Context, order *entity.Order) error {
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

// CreatePayment creates a payment record for the order.
func (a *orderActivities) CreatePayment(
	ctx context.Context,
	order *entity.Order,
) (uuid.UUID, error) {
	a.logger.Infof("Create payment for order: %s", order.ID)

	// Step 1: Create and publish payment request event
	err := a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create payment request event
		paymentEvent := producer.NewPaymentRequestEvent(
			order.ID,
			order.CustomerID,
			order.TotalPrice,
			order.Currency,
			order.PaymentGateway,
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
		constant.CreatePaymentStepTimeout,
	)
	if err != nil {
		a.logger.Errorf("Failed to receive payment response for order %s: %v", order.ID, err)

		return uuid.Nil, fmt.Errorf("failed to receive payment response: %w", err)
	}

	a.logger.Infof("Successfully received payment response for order %s: ID=%s, Status=%s",
		order.ID, response.PaymentID, response.Status)

	return response.PaymentID, nil
}

// SendPaymentRequiredNotification sends a payment required notification to the customer.
func (a *orderActivities) SendPaymentRequiredNotification(
	ctx context.Context,
	order *entity.Order,
	reservedProducts []entity.Product,
	customerEmail string,
) error {
	a.logger.Infof("Sending payment required notification for order: %s", order.ID)

	err := a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		notificationEvent := producer.NewNotificationRequestEvent(
			order,
			reservedProducts,
			customerEmail,
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
			AggregateID:   order.ID,
			EventType:     kafka.NotificationRequestedEventType,
			Topic:         kafka.NotificationRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			a.logger.Errorf(
				"Failed to create payment required notification for order %s: %v",
				order.ID,
				err,
			)

			return fmt.Errorf("failed to create payment required notification event: %w", err)
		}

		a.logger.Infof(
			"Successfully created payment required notification for order: %s",
			order.ID,
		)

		return nil
	})
	if err != nil {
		a.logger.Errorf(
			"Failed to create payment required notification for order %s: %v",
			order.ID,
			err,
		)

		return fmt.Errorf("failed to create payment required notification: %w", err)
	}

	return nil
}

// WaitForPaymentConfirmation waits for payment confirmation, with payment reminder notification, no retry and auto-cancel.
func (a *orderActivities) WaitForPaymentConfirmation(
	ctx context.Context,
	order *entity.Order,
	customerEmail string,
) (uuid.UUID, error) {
	a.logger.Infof("Waiting for payment confirmation for order: %s", order.ID)

	taskIDs, scheduleErr := a.schedulePaymentReminders(ctx, order, customerEmail)
	if scheduleErr != nil {
		a.logger.Errorf(
			"Failed to schedule payment reminders for order %s: %v",
			order.ID,
			scheduleErr,
		)
	} else {
		a.logger.Infof("Scheduled %d payment reminder tasks for order %s: %v", len(taskIDs), order.ID, taskIDs)
	}

	response, err := a.paymentClient.WaitForPaymentResponse(
		ctx,
		order.ID,
		constant.WaitForPaymentConfirmationStepTimeout,
	)
	// Cancel payment reminders after payment completion or timeout
	if len(taskIDs) > 0 {
		if cleanupErr := a.taskCancellationHelper.CancelPaymentReminderTasksByIDs(ctx, taskIDs); cleanupErr != nil {
			a.logger.Errorf(
				"Failed to cancel payment reminders for order %s: %v",
				order.ID,
				cleanupErr,
			)
		}
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to receive payment confirmation: %w", err)
	}

	a.logger.Infof("Successfully received payment confirmation for order %s: ID=%s, Status=%s",
		order.ID, response.PaymentID, response.Status)

	return response.PaymentID, nil
}

// createPaymentReminderRequest creates a payment reminder request.
func (a *orderActivities) createPaymentReminderRequest(
	order *entity.Order,
	customerEmail string,
	correlationID uuid.UUID,
	reminderCount int,
) *dto.PaymentReminderRequest {
	return &dto.PaymentReminderRequest{
		OrderID:       order.ID,
		CorrelationID: correlationID,
		CustomerEmail: customerEmail,
		PaymentID:     uuid.Nil, // Will be set by payment service
		TotalPrice:    order.TotalPrice,
		Currency:      order.Currency,
		ReminderCount: reminderCount,
	}
}

// schedulePaymentTask schedules a payment reminder task and returns the task ID.
func (a *orderActivities) schedulePaymentTask(
	req *dto.PaymentReminderRequest,
	delay time.Duration,
	taskType string,
) (string, error) {
	reminderTask, err := task.NewPaymentReminderTask(req)
	if err != nil {
		return "", fmt.Errorf("failed to create %s reminder task: %w", taskType, err)
	}

	taskInfo, err := a.asynqClient.EnqueueIn(delay, reminderTask)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue %s reminder: %w", taskType, err)
	}

	return taskInfo.ID, nil
}

// schedulePaymentReminders schedules payment reminder tasks using asynq.
func (a *orderActivities) schedulePaymentReminders(
	_ context.Context,
	order *entity.Order,
	customerEmail string,
) ([]string, error) {
	a.logger.Infof("Scheduling payment reminders for order: %s", order.ID)

	correlationID := uuid.New()
	taskIDs := make([]string, 0)

	// Schedule first reminder
	firstReq := a.createPaymentReminderRequest(
		order,
		customerEmail,
		correlationID,
		constant.FirstReminderSequence,
	)

	firstTaskID, err := a.schedulePaymentTask(firstReq, constant.FirstPaymentReminderDelay, "first")
	if err != nil {
		return nil, err
	}

	taskIDs = append(taskIDs, firstTaskID)

	// Schedule second reminder
	secondReq := a.createPaymentReminderRequest(
		order,
		customerEmail,
		correlationID,
		constant.SecondReminderSequence,
	)

	secondTaskID, err := a.schedulePaymentTask(
		secondReq,
		constant.SecondPaymentReminderDelay,
		"second",
	)
	if err != nil {
		return nil, err
	}

	taskIDs = append(taskIDs, secondTaskID)

	// Schedule order expiration
	expireTask, err := task.NewExpireOrderPaymentTask(dto.ExpireOrderPaymentRequest{
		CustomerID:     order.CustomerID,
		CustomerEmail:  customerEmail,
		OrderID:        order.ID,
		CorrelationID:  correlationID,
		IdempotencyKey: uuid.New(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create expiration task: %w", err)
	}

	expireTaskInfo, err := a.asynqClient.EnqueueIn(constant.ExpireOrderReminderDelay, expireTask)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue order expiration: %w", err)
	}

	taskIDs = append(taskIDs, expireTaskInfo.ID)

	a.logger.Infof(
		"Successfully scheduled payment reminders and expiration for order: %s with task IDs: %v",
		order.ID,
		taskIDs,
	)

	return taskIDs, nil
}

// ConfirmProductsDeduction confirms stock deduction after successful payment.
func (a *orderActivities) ConfirmProductsDeduction(
	ctx context.Context,
	order *entity.Order,
	_ []entity.Product,
) error {
	a.logger.Infof(
		"Confirming inventory deduction for order: %s",
		order.ID,
	)

	// Log deduction details for each item
	for i := range order.Items {
		item := &order.Items[i]
		a.logger.Infof("Confirming deduction of %d units of product %s for order %s",
			item.Quantity, item.ProductID, order.ID)
	}

	deductionItems := make([]dto.ProductRestorationItem, len(order.Items))

	for i := range order.Items {
		deductionItems[i] = dto.ProductRestorationItem{
			ProductID: order.Items[i].ProductID,
			Quantity:  order.Items[i].Quantity,
		}
	}

	// Add user authentication info to context for gRPC calls
	ctx = addUserAuthToContext(ctx, a.logger, order.ID)

	_, err := a.productClient.ConfirmProductsDeduction(ctx, deductionItems)
	if err != nil {
		a.logger.Errorf("Failed to confirm stock deduction for order %s: %v", order.ID, err)
		sagaErr := CategorizeError(constant.ConfirmProductsDeductionStep, err)

		return sagaErr
	}

	a.logger.Infof("Successfully confirmed stock deduction for order: %s", order.ID)

	return nil
}

// ProcessFulfillment creates shipping/fulfillment arrangement for the order.
func (a *orderActivities) ProcessFulfillment(
	ctx context.Context,
	payload *Payload,
) (uuid.UUID, decimal.Decimal, string, error) {
	a.logger.Infof("Creating shipping for order: %s", payload.Order.ID)

	err := a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create fulfillment request event
		fulfillmentEvent := producer.NewFulfillmentRequestEvent(payload.Order, &payload.Shipping)

		evtPayload, err := json.Marshal(fulfillmentEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal fulfillment request event: %w", err)
		}

		// Create outbox event for reliable delivery
		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "fulfillment",
			AggregateID:   payload.Order.ID,
			EventType:     kafka.FulfillmentRequestedEventType,
			Topic:         kafka.FulfillmentRequestTopic,
			Payload:       evtPayload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
			return fmt.Errorf("failed to create fulfillment request event: %w", err)
		}

		a.logger.Infof("Successfully created fulfillment request for order: %s", payload.Order.ID)

		return nil
	})
	if err != nil {
		a.logger.Errorf(
			"Failed to publish fulfillment request for order %s: %v",
			payload.Order.ID,
			err,
		)

		return uuid.Nil, decimal.Zero, "", fmt.Errorf("failed to create shipping: %w", err)
	}

	// Step 2: Wait for fulfillment service response
	a.logger.Infof("Waiting for fulfillment response for order: %s", payload.Order.ID)

	response, err := a.fulfillmentClient.WaitForFulfillmentResponse(
		ctx,
		payload.Order.ID,
		constant.ProcessFulfillmentStepTimeout,
	)
	if err != nil {
		a.logger.Errorf(
			"Failed to receive fulfillment response for order %s: %v",
			payload.Order.ID,
			err,
		)

		return uuid.Nil, decimal.Zero, "", fmt.Errorf(
			"failed to receive fulfillment response: %w",
			err,
		)
	}

	a.logger.Infof(
		"Successfully received fulfillment response for order %s: ID=%s, ShippingCost=%s, Tracking=%s",
		payload.Order.ID,
		response.FulfillmentID,
		response.ShippingCost,
		response.TrackingNumber,
	)

	return response.FulfillmentID, response.ShippingCost, response.TrackingNumber, nil
}

// SendOrderConfirmedNotification sends order confirmation to customer.
func (a *orderActivities) SendOrderConfirmedNotification(
	ctx context.Context,
	order *entity.Order,
	products []entity.Product,
	trackingNumber *string,
	customerEmail string,
) error {
	a.logger.Infof(
		"Sending order confirmation for order: %s with tracking: %s to email: %s",
		order.ID,
		trackingNumber,
		customerEmail,
	)

	// Create notification request event
	err := a.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		outboxRepo := ds.OutboxRepository()

		// Create order confirmation notification event
		notificationEvent := producer.NewNotificationRequestEvent(
			order,
			products,
			customerEmail,
			"Customer Name",
			trackingNumber,
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
			AggregateID:   order.ID,
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

		a.logger.Infof(
			"Successfully created order confirmation notification for order: %s",
			order.ID,
		)

		return nil
	})
	if err != nil {
		a.logger.Errorf("Failed to create notification request for order %s: %v", order.ID, err)

		return fmt.Errorf("failed to send order confirmation: %w", err)
	}

	a.logger.Infof(
		"Order confirmation notification queued for order: %s with tracking: %s",
		order.ID,
		trackingNumber,
	)

	return nil
}
