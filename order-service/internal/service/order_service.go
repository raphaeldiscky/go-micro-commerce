// Package service provides business logic for order operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/pageutils"

	pkgDto "github.com/raphaeldiscky/go-micro-template/pkg/dto"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/utils/redisutils"
)

// OrderServiceInterface defines the interface for order business operations.
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
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
	PayOrder(ctx context.Context, req dto.PayOrderRequest, id uuid.UUID) (*dto.OrderResponse, error)
}

// OrderService implements the OrderServiceInterface.
type OrderService struct {
	dataStore              repository.DataStore
	logger                 logger.Logger
	orderLifecycleProducer mq.KafkaProducerInterface
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	orderLifecycleProducer mq.KafkaProducerInterface,
) OrderServiceInterface {
	return &OrderService{
		dataStore:              dataStore,
		logger:                 appLogger,
		orderLifecycleProducer: orderLifecycleProducer,
	}
}

//nolint:gocyclo,revive,cyclop // ignore complexity, CreateOrder is large but intentional
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
		productRepo := ds.ProductRepository()
		outboxRepo := ds.OutboxRepository()

		order, err := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return err
		}

		if order != nil && order.CustomerID == req.CustomerID {
			res = dto.MapToOrderResponse(order)

			return nil
		}

		productIDs := make([]uuid.UUID, len(req.Items))
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
		}

		products, err := productRepo.FindByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return err
		}

		if len(products) != len(productIDs) {
			return httperror.NewInternalServerError("failed to get all products")
		}

		var orderItems []entity.OrderItem

		for i, product := range products {
			if product.Quantity < req.Items[i].Quantity {
				return httperror.NewInsufficientProductStockError()
			}

			product.Quantity -= req.Items[i].Quantity
			orderItem := entity.OrderItem{
				ID:        uuid.New(),
				ProductID: product.ID,
				Quantity:  req.Items[i].Quantity,
				Price:     product.Price,
			}
			orderItems = append(orderItems, orderItem)
		}

		if err := productRepo.BulkUpdateQuantity(ctx, products); err != nil {
			return err
		}

		// Create domain entity
		newOrder, err := entity.NewOrder(req.CustomerID, req.IdempotencyKey, orderItems)
		if err != nil {
			return err
		}

		// Save to repository
		savedOrder, err := orderRepo.Create(ctx, newOrder)
		if err != nil {
			return err
		}

		// Publish domain event
		evt := event.NewOrderLifecycleEvent(
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
			EventType:     constant.KafkaEventTypeOrderCreated,
			Topic:         constant.TopicOrderLifecycle,
			Payload:       payload,
			Status:        constant.OutboxStatusPending,
			CreatedAt:     time.Now().UTC(),
			ScheduledFor:  time.Now().UTC(),
			Attempts:      0,
		}

		if err := outboxRepo.Create(ctx, outboxEvent); err != nil {
			return err
		}

		res = dto.MapToOrderResponse(savedOrder)

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

	return dto.MapToOrderResponse(order), nil
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
		res[i] = *dto.MapToOrderResponse(order)
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
		res[i] = *dto.MapToOrderResponse(order)
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
		evt := event.NewOrderLifecycleEvent(
			updatedOrder.ID,
			status,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order status updated event")
		}

		res = dto.MapToOrderResponse(updatedOrder)

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
		evt := event.NewOrderLifecycleEvent(
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

// PayOrder processes payment for an order.
func (s *OrderService) PayOrder(
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

		updatedOrder := existingOrder
		updatedOrder.IdempotencyKey = req.IdempotencyKey
		// Update status to paid
		if err := updatedOrder.UpdateStatus(constant.OrderStatusPaid); err != nil {
			return httperror.NewBadRequestError("failed to update order status entity")
		}

		// Set idempotency key
		updatedOrder.IdempotencyKey = req.IdempotencyKey

		s.logger.Infof("---5 passes: %+v", updatedOrder)

		// Save updated order
		updatedOrder, err = orderRepo.Update(ctx, updatedOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to update order")
		}

		s.logger.Infof("existing key: %+v", existingOrder.IdempotencyKey)
		s.logger.Infof("updated key: %+v", updatedOrder.IdempotencyKey)

		s.logger.Infof("---6 passes: %+v", req)

		// Publish domain event
		evt := event.NewOrderLifecycleEvent(
			updatedOrder.ID,
			constant.OrderStatusPaid,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
			updatedOrder.Items,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order paid event")
		}

		res = dto.MapToOrderResponse(updatedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
