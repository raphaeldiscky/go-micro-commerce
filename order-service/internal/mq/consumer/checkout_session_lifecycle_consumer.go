package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/shopspring/decimal"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// CheckoutSessionLifecycleEvent is the envelope for checkout session lifecycle events.
type CheckoutSessionLifecycleEvent struct {
	Metadata kafkaevent.Metadata                          `json:"metadata"`
	Payload  kafkaevent.CheckoutSessionOrderPlacedPayload `json:"payload"`
}

// CheckoutSessionLifecycleConsumer handles the logic for processing checkout session lifecycle events.
type CheckoutSessionLifecycleConsumer struct {
	logger           logger.Logger
	datastore        repository.DataStore
	sagaOrchestrator saga.Orchestrator
}

// NewCheckoutSessionLifecycleConsumer creates a new consumer for checkout session lifecycle events.
func NewCheckoutSessionLifecycleConsumer(
	appLogger logger.Logger,
	ds repository.DataStore,
	sagaOrchestrator saga.Orchestrator,
) *CheckoutSessionLifecycleConsumer {
	return &CheckoutSessionLifecycleConsumer{
		logger:           appLogger,
		datastore:        ds,
		sagaOrchestrator: sagaOrchestrator,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *CheckoutSessionLifecycleConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata kafkaevent.Metadata `json:"metadata"`
	}

	if err := json.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	// Store event in inbox for idempotent processing
	inboxEvent := entity.NewInboxEvent(
		meta.Metadata.EventID,
		"checkout_session", // aggregate type
		meta.Metadata.AggregateID,
		meta.Metadata.EventType,
		kafka.CheckoutSessionLifecycleTopic, // topic
		meta.Metadata.Source,
		json.RawMessage(body),
		nil, // correlation_id - could be extracted from metadata if available
		nil, // causation_id - could be extracted from metadata if available
	)

	return c.datastore.Atomic(ctx, func(ds repository.DataStore) error {
		inboxRepo := ds.InboxRepository()

		// Store event in inbox (handles duplicates automatically)
		storedEvent, err := inboxRepo.Create(ctx, inboxEvent)
		if err != nil {
			return fmt.Errorf("failed to store event in inbox: %w", err)
		}

		// If it's a duplicate, just log and return successfully
		if storedEvent.Status == constant.InboxStatusDuplicate {
			c.logger.Infof(
				"Duplicate event received: %s, skipping processing",
				meta.Metadata.EventID,
			)

			return nil
		}

		// Mark as processing
		if err = inboxRepo.MarkAsProcessing(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processing: %w", err)
		}

		// Process the event based on type
		var processingErr error

		switch meta.Metadata.EventType {
		case kafka.CheckoutSessionOrderPlacedEventType:
			processingErr = c.processCheckoutSessionOrderPlaced(ctx, ds, body)
		case kafka.CheckoutSessionCanceledEventType:
			processingErr = c.processCheckoutSessionCanceled(ctx, ds, body)
		default:
			c.logger.Warnf("ignoring unknown event type: %s", meta.Metadata.EventType)
			// Mark as processed even for unknown events to avoid reprocessing
			processingErr = nil
		}

		// Update inbox event status based on processing result
		if processingErr != nil {
			c.logger.Errorf("Failed to process event %s: %v", meta.Metadata.EventID, processingErr)

			if err = inboxRepo.MarkAsFailed(ctx, storedEvent.ID, processingErr.Error()); err != nil {
				return fmt.Errorf("failed to mark event as failed: %w", err)
			}

			return processingErr
		}

		if err = inboxRepo.MarkAsProcessed(ctx, storedEvent.ID); err != nil {
			return fmt.Errorf("failed to mark event as processed: %w", err)
		}

		return nil
	})
}

// processCheckoutSessionOrderPlaced handles checkout session placed order events.
// This creates an order and triggers the saga orchestration.
func (c *CheckoutSessionLifecycleConsumer) processCheckoutSessionOrderPlaced(
	ctx context.Context,
	ds repository.DataStore,
	body []byte,
) error {
	var evt CheckoutSessionLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session placed order event: %w", err)
	}

	c.logger.Infof(
		"Handling checkout session placed order event for checkout session ID: %s",
		evt.Payload.CheckoutSessionID,
	)

	// Validate checkout session ID is not nil
	if evt.Payload.CheckoutSessionID == uuid.Nil {
		return errors.New("checkout session ID is nil in event payload")
	}

	c.logger.Debugf(
		"Event payload - CheckoutSessionID: %s, IdempotencyKey: %s, UserID: %s",
		evt.Payload.CheckoutSessionID,
		evt.Payload.IdempotencyKey,
		evt.Payload.UserID,
	)

	orderRepo := ds.OrderRepository()

	// Check for existing order with same idempotency key (idempotent processing)
	existingOrder, err := orderRepo.FindByIdempotencyKey(ctx, evt.Payload.IdempotencyKey)
	if err != nil && err.Error() != constant.OrderNotFoundErrorMessage {
		return fmt.Errorf("failed to check existing order: %w", err)
	}

	// If order already exists for this customer, skip creation (idempotent)
	if existingOrder != nil && existingOrder.CustomerID == evt.Payload.UserID {
		c.logger.Infof(
			"Order %s already exists for idempotency key, skipping creation (idempotent)",
			existingOrder.ID,
		)

		// Trigger saga if needed (saga execution is also idempotent)
		userAuth := pkgdto.UserAuthInfo{
			UserID: evt.Payload.UserID,
			Email:  "system@order-service.internal", // Placeholder for event-driven saga
			Roles:  []string{"user"},                // Default role
		}
		sagaCtx := echoutils.AddUserAuthToContexts(ctx, userAuth)
		payload := &saga.Payload{Order: existingOrder}
		c.sagaOrchestrator.ExecuteOrderSagaAsync(sagaCtx, payload)

		c.logger.Infof(
			"Triggered saga for existing order %s from checkout session %s",
			existingOrder.ID,
			evt.Payload.CheckoutSessionID,
		)

		return nil // Success - idempotent behavior
	}

	// Create order items from checkout session items
	orderItems := make([]entity.OrderItem, len(evt.Payload.Items))
	for i := range evt.Payload.Items {
		item := evt.Payload.Items[i]
		orderItems[i] = entity.OrderItem{
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TaxRate:       decimal.Zero,
			TotalTax:      decimal.Zero,
			TotalDiscount: decimal.Zero,
			TotalPrice:    item.UnitPrice.Mul(decimal.NewFromInt(item.Quantity)),
		}
	}

	// Map shipping data from event to entity
	courier := entity.Courier{
		CourierID: evt.Payload.Courier.CourierID,
	}

	destination := entity.Destination{
		City:        evt.Payload.Destination.City,
		State:       evt.Payload.Destination.State,
		PostalCode:  evt.Payload.Destination.PostalCode,
		CountryCode: evt.Payload.Destination.CountryCode,
	}

	origin := entity.Origin{
		City:        evt.Payload.Origin.City,
		State:       evt.Payload.Origin.State,
		PostalCode:  evt.Payload.Origin.PostalCode,
		CountryCode: evt.Payload.Origin.CountryCode,
	}

	packageData := entity.Package{
		WeightKG: evt.Payload.Package.WeightKG,
		Width:    evt.Payload.Package.Width,
		Height:   evt.Payload.Package.Height,
		Length:   evt.Payload.Package.Length,
		Unit:     evt.Payload.Package.Unit,
	}

	// Create order entity with pre-calculated values from cart-service
	order, err := entity.NewOrder(
		evt.Payload.UserID,
		evt.Payload.IdempotencyKey,
		evt.Payload.CheckoutSessionID,
		constant.PaymentGateway(evt.Payload.PaymentGateway),
		evt.Payload.Currency,
		courier,
		destination,
		origin,
		packageData,
		orderItems,
	)
	if err != nil {
		return fmt.Errorf("failed to create order entity: %w", err)
	}

	// Set pre-calculated shipping cost and total amount from cart-service
	order.ShippingCost = evt.Payload.ShippingCost
	order.TotalPrice = evt.Payload.TotalAmount

	// Calculate and set subtotal from items (for consistency)
	subtotal := decimal.Zero
	for _, item := range orderItems {
		subtotal = subtotal.Add(item.TotalPrice)
	}

	order.Subtotal = subtotal

	c.logger.Debugf(
		"Created order entity - OrderID: %s, CheckoutSessionID: %s, CustomerID: %s",
		order.ID,
		order.CheckoutSessionID,
		order.CustomerID,
	)

	// Save order to database
	createdOrder, err := orderRepo.Create(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	c.logger.Infof(
		"Created order %s from checkout session %s",
		createdOrder.ID,
		evt.Payload.CheckoutSessionID,
	)

	// Create user auth context for saga execution from event payload
	userAuth := pkgdto.UserAuthInfo{
		UserID: evt.Payload.UserID,
		Email:  "system@order-service.internal", // Placeholder for event-driven saga
		Roles:  []string{"user"},                // Default role
	}

	// Add user auth to context for async saga execution
	sagaCtx := echoutils.AddUserAuthToContexts(ctx, userAuth)

	// Create saga payload
	payload := &saga.Payload{
		Order: createdOrder,
	}

	c.logger.Debugf("Saga payload: %+v", payload)

	// Trigger order saga asynchronously with user auth context
	c.sagaOrchestrator.ExecuteOrderSagaAsync(sagaCtx, payload)

	c.logger.Infof(
		"Successfully triggered order creation saga for checkout session ID: %s, Order ID: %s",
		evt.Payload.CheckoutSessionID,
		createdOrder.ID,
	)

	return nil
}

// processCheckoutSessionCanceled handles checkout session canceled events.
func (c *CheckoutSessionLifecycleConsumer) processCheckoutSessionCanceled(
	_ context.Context,
	_ repository.DataStore,
	body []byte,
) error {
	var evt CheckoutSessionLifecycleEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal checkout session canceled event: %w", err)
	}

	c.logger.Infof(
		"Handling checkout session canceled event for checkout session ID: %s",
		evt.Payload.CheckoutSessionID,
	)

	// For canceled checkout sessions, we don't need to do anything in order-service
	// The checkout session was canceled before an order was created
	c.logger.Infof(
		"Checkout session %s was canceled, no order created",
		evt.Payload.CheckoutSessionID,
	)

	return nil
}
