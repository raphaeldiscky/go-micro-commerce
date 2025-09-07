// Package service provides business logic for order operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"
	"github.com/shopspring/decimal"

	pkgDto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mq/producer"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/redisutils"
)

// OrderServiceInterface defines the interface for order business operations.
type OrderServiceInterface interface {
	CreateOrder(
		ctx context.Context,
		req dto.CreateOrderRequest,
	) (*dto.OrderResponse, error)
	CreateOrderWithSaga(
		ctx context.Context,
		req dto.CreateOrderRequest,
	) (*dto.OrderResponse, error)
	CreateOrderWithTemporal(
		ctx context.Context,
		req dto.CreateOrderRequest,
	) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	GetOrdersByCustomer(
		ctx context.Context,
		customerID uuid.UUID,
		req dto.GetOrdersRequest,
	) ([]dto.OrderResponse, *pkgDto.PageMetaData, error)
	GetOrders(
		ctx context.Context,
		req dto.GetOrdersRequest,
	) ([]dto.OrderResponse, *pkgDto.PageMetaData, error)
	UpdateOrderStatus(
		ctx context.Context,
		id uuid.UUID,
		status constant.OrderStatus,
	) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, req dto.CancelOrderRequest, id uuid.UUID) error
	NotifyOrderFailure(
		ctx context.Context,
		orderID uuid.UUID,
		status constant.OrderStatus,
		reason string,
	) error
	RequestPaymentOrder(
		ctx context.Context,
		req dto.PayOrderRequest,
		id uuid.UUID,
	) (*dto.OrderResponse, error)
}

// OrderService implements the OrderServiceInterface.
type OrderService struct {
	dataStore              repository.DataStore
	productClient          client.ProductClientInterface
	logger                 logger.Logger
	orderLifecycleProducer kafka.ProducerInterface
	sagaOrchestrator       saga.Orchestrator
	temporalClient         *client.TemporalClient
	config                 *config.Config
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(
	cfg *config.Config,
	dataStore repository.DataStore,
	productClient client.ProductClientInterface,
	appLogger logger.Logger,
	orderLifecycleProducer kafka.ProducerInterface,
	sagaOrchestrator saga.Orchestrator,
	temporalClient *client.TemporalClient,
) OrderServiceInterface {
	return &OrderService{
		dataStore:              dataStore,
		productClient:          productClient,
		logger:                 appLogger,
		orderLifecycleProducer: orderLifecycleProducer,
		sagaOrchestrator:       sagaOrchestrator,
		temporalClient:         temporalClient,
		config:                 cfg,
	}
}

//nolint:gocyclo,revive,cyclop,funlen,gocognit // ignore complexity, CreateOrder is large but intentional
func (s *OrderService) CreateOrder(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	res := new(dto.OrderResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		outboxRepo := ds.OutboxRepository()

		order, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return err
		}

		if order != nil && order.CustomerID == req.CustomerID {
			res = mapper.MapToOrderResponse(order)

			return nil
		}

		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		if s.productClient == nil {
			return httperror.NewServiceUnavailableError("product service is currently unavailable")
		}

		// First, get products to check availability and get current versions
		products, err := s.productClient.GetProducts(ctx, productIDs)
		if err != nil {
			return err
		}

		s.logger.Infof("====1==== products: %v", products)

		for i, product := range products {
			s.logger.Infof(
				"====1.%d==== product: ID=%s, Quantity=%d, Version=%d, ReservedQuantity=%d",
				i+1,
				product.ID,
				product.Quantity,
				product.Version,
				product.ReservedQuantity,
			)
		}

		if len(products) != len(productIDs) {
			return httperror.NewInternalServerError("failed to get all products")
		}

		// Create product map for quick lookup
		productMap := make(map[uuid.UUID]*entity.Product)
		for i, product := range products {
			productMap[product.ID] = &products[i]
		}

		// Create reservations with expected versions from fetched products
		reservations := make([]dto.ProductReservationItem, len(req.Items))

		for i, item := range req.Items {
			product, exists := productMap[item.ProductID]
			if !exists {
				return httperror.NewBadRequestError("product not found")
			}

			reservations[i] = dto.ProductReservationItem{
				ProductID:       item.ProductID,
				Quantity:        item.Quantity,
				ExpectedVersion: product.Version,
			}
		}

		s.logger.Infof("====2==== reservations: %v", reservations)

		reservedProducts, err := s.productClient.ReserveProducts(
			ctx,
			req.IdempotencyKey,
			reservations,
		)
		if err != nil {
			s.logger.Errorf("====ERROR==== ReserveProducts failed: %v", err)

			return err
		}

		var orderItems []entity.OrderItem

		for i, product := range reservedProducts {
			orderItem, err := entity.NewOrderItem(
				product.ID,
				req.Items[i].Quantity,
				product.UnitPrice,
			)
			if err != nil {
				return err
			}

			orderItems = append(orderItems, *orderItem)
		}

		// Create domain entity
		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, "IDR", orderItems)
		if err != nil {
			return err
		}

		s.logger.Infof("====3==== newOrder: %v", newOrder)

		// Save to repository
		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			return err
		}

		// Update reservations with current versions from reserved products
		updatedReservations := make([]dto.ProductReservationItem, len(reservedProducts))
		for i, product := range reservedProducts {
			updatedReservations[i] = dto.ProductReservationItem{
				ProductID:       product.ID,
				Quantity:        reservations[i].Quantity,
				ExpectedVersion: product.Version, // Use the updated version from reservation
			}
		}

		s.logger.Infof("====4==== updatedReservations: %v", updatedReservations)

		// Confirm product reservations
		_, err = s.productClient.ConfirmProductsDeduction(ctx, updatedReservations)
		if err != nil {
			return err
		}

		s.logger.Infof("====5==== savedOrder: %v", savedOrder)

		// Publish domain event
		evt := producer.NewOrderLifecycleEvent(
			savedOrder.ID,
			constant.OrderStatusPending,
			savedOrder.CustomerID,
			savedOrder.TotalPrice,
			savedOrder.Items,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal order event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "order",
			AggregateID:   savedOrder.ID,
			EventType:     kafka.OrderCreatedEventType,
			Topic:         kafka.OrderLifecycleTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return err
		}

		res = mapper.MapToOrderResponse(savedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CreateOrderWithTemporal handles POST /orders/temporal with Temporal processing.
func (s *OrderService) CreateOrderWithTemporal(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	s.logger.Infof("Creating order with Temporal workflow for customer: %s", req.CustomerID)

	if s.temporalClient == nil {
		return nil, httperror.NewServiceUnavailableError("Temporal service is not available")
	}

	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	var res *dto.OrderResponse

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		existingOrder, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return err
		}

		if existingOrder != nil && existingOrder.CustomerID == req.CustomerID {
			res = mapper.MapToOrderResponse(existingOrder)

			return nil
		}

		orderItems := make([]entity.OrderItem, len(req.Items))
		for i := range req.Items {
			orderItems[i] = entity.OrderItem{
				ID:        uuid.New(),
				ProductID: req.Items[i].ProductID,
				Quantity:  req.Items[i].Quantity,
				UnitPrice: decimal.Zero, // Will be set by Temporal workflows
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
		}

		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, "IDR", orderItems)
		if err != nil {
			return fmt.Errorf("failed to create order entity: %w", err)
		}

		if err := newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			return err
		}

		// Start Temporal workflow
		req := dto.TemporalOrderSagaRequest{
			Order: savedOrder,
		}

		workflowOptions := s.temporalClient.CreateWorkflowOptions(savedOrder.ID)

		workflowRun, err := s.temporalClient.Client.ExecuteWorkflow(
			ctx,
			workflowOptions,
			constant.OrderSagaWorkflowName,
			req,
		)
		if err != nil {
			s.logger.Errorf(
				"Failed to start Temporal workflow for order %s: %v",
				savedOrder.ID,
				err,
			)

			return fmt.Errorf("failed to start order processing workflow: %w", err)
		}

		s.logger.Infof("Started Temporal workflow for order %s with workflow ID: %s",
			savedOrder.ID, workflowRun.GetID())

		res = mapper.MapToOrderResponse(savedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetOrder retrieves an order by ID.
func (s *OrderService) GetOrder(
	ctx context.Context,
	id uuid.UUID,
) (*dto.OrderResponse, error) {
	orderRepo := s.dataStore.OrderRepository()

	order, err := orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get order")
	}

	if order == nil {
		return nil, httperror.NewOrderNotFoundError()
	}

	return mapper.MapToOrderResponse(order), nil
}

// GetOrdersByCustomer retrieves orders for a specific customer with pagination.
func (s *OrderService) GetOrdersByCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	req dto.GetOrdersRequest,
) ([]dto.OrderResponse, *pkgDto.PageMetaData, error) {
	var orders []*entity.Order

	var total int64

	var err error

	orderRepo := s.dataStore.OrderRepository()
	offset := pageutils.GetOffset(req.Page, req.Limit)

	orders, err = orderRepo.FindByCustomerID(ctx, customerID, req.Limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get customer orders")
	}

	res := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		res[i] = *mapper.MapToOrderResponse(order)
	}

	total, err = orderRepo.CountByCustomer(ctx, customerID)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to count customer orders")
	}

	metadata := pageutils.NewMetadata(total, req.Page, req.Limit)

	return res, metadata, nil
}

// GetOrders retrieves all orders with pagination.
func (s *OrderService) GetOrders(
	ctx context.Context,
	req dto.GetOrdersRequest,
) ([]dto.OrderResponse, *pkgDto.PageMetaData, error) {
	var orders []*entity.Order

	var total int64

	var err error

	orderRepo := s.dataStore.OrderRepository()
	offset := pageutils.GetOffset(req.Page, req.Limit)

	orders, err = orderRepo.FindAll(ctx, req.Limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get orders")
	}

	res := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		res[i] = *mapper.MapToOrderResponse(order)
	}

	total, err = orderRepo.Count(ctx)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to count orders")
	}

	metadata := pageutils.NewMetadata(total, req.Page, req.Limit)

	return res, metadata, nil
}

// UpdateOrderStatus updates only the status of an order.
func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.OrderStatus,
) (*dto.OrderResponse, error) {
	res := new(dto.OrderResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check if order exists
		existingOrder, err := orderRepo.FindByID(ctx, id)
		if err != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Update status
		if err := existingOrder.UpdateStatus(status); err != nil {
			return httperror.NewBadRequestError("invalid order status")
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, existingOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to update order status")
		}

		// Publish domain event
		evt := producer.NewOrderLifecycleEvent(
			updatedOrder.ID,
			status,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order status updated event")
		}

		res = mapper.MapToOrderResponse(updatedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CancelOrder cancels an order.
func (s *OrderService) CancelOrder(
	ctx context.Context,
	req dto.CancelOrderRequest,
	id uuid.UUID,
) error {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return err
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check if order exists
		existingOrder, err := orderRepo.FindByID(ctx, id)
		if err != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Check if order can be canceled
		if !existingOrder.CanBeCancelled() {
			return httperror.NewBadRequestError("order cannot be canceled in current status")
		}

		if existingOrder.IdempotencyKey == req.IdempotencyKey {
			return httperror.NewBadRequestError(
				fmt.Sprintf(
					"idempotency key for update conflict, existing key: %s , updated key: %s",
					existingOrder.IdempotencyKey,
					req.IdempotencyKey,
				),
			)
		}

		// Update status to canceled
		if err := existingOrder.UpdateStatus(constant.OrderStatusCanceled); err != nil {
			return httperror.NewBadRequestError("failed to cancel order entity")
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, existingOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to cancel order")
		}

		// Publish domain event
		evt := producer.NewOrderLifecycleEvent(
			existingOrder.ID,
			constant.OrderStatusCanceled,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order canceled event")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RequestPaymentOrder initiates payment processing for an order by publishing a payment request producer.
func (s *OrderService) RequestPaymentOrder(
	ctx context.Context,
	req dto.PayOrderRequest,
	id uuid.UUID,
) (*dto.OrderResponse, error) {
	lockRepo := s.dataStore.LockRepository()
	lockKey := redisutils.NewLockKey(req.IdempotencyKey, req.CustomerID)
	ttl := constant.CreateOrderTTL
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(constant.CreateOrderRetryInterval),
			constant.CreateOrderRetryLimit,
		),
	}

	lock, err := lockRepo.Get(ctx, lockKey, ttl, opt)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	res := new(dto.OrderResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if order exists
		existingOrder, err := orderRepo.FindByID(ctx, id)
		if err != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Check if order can be paid
		if !existingOrder.CanBePaid() {
			return httperror.NewBadRequestError("order cannot be paid in current status")
		}

		if existingOrder.IdempotencyKey == req.IdempotencyKey {
			return httperror.NewBadRequestError(
				fmt.Sprintf(
					"idempotency key for update conflict, existing key: %s , updated key: %s",
					existingOrder.IdempotencyKey,
					req.IdempotencyKey,
				),
			)
		}

		// Update idempotency key but keep order in pending status
		// Payment service will update the order status when payment completes
		updatedOrder := existingOrder
		updatedOrder.IdempotencyKey = req.IdempotencyKey

		// Save updated order
		updatedOrder, err = orderRepo.Update(ctx, updatedOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to update order")
		}

		// Publish payment request event to payment service
		evt := producer.NewPaymentRequestEvent(
			updatedOrder.ID,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			"IDR", // Default currency
			req.PaymentMethod,
		)

		payload, err := json.Marshal(evt)
		if err != nil {
			return httperror.NewInternalServerError("failed to marshal payment request event")
		}

		outboxEvent := &entity.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "payment",
			AggregateID:   updatedOrder.ID,
			EventType:     kafka.PaymentRequestedEventType,
			Topic:         kafka.PaymentRequestTopic,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return httperror.NewInternalServerError("failed to create payment request event")
		}

		res = mapper.MapToOrderResponse(updatedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// NotifyOrderFailure updates the order status to failed and logs the reason.
func (s *OrderService) NotifyOrderFailure(
	_ context.Context,
	orderID uuid.UUID,
	status constant.OrderStatus,
	reason string,
) error {
	if status != constant.OrderStatusFailed {
		return nil
	}

	s.logger.Infof("Send order failure notification: %s, reason: %s", orderID, reason)

	return nil
}
