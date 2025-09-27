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
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"
	"github.com/shopspring/decimal"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

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

// OrderService defines the interface for order business operations.
type OrderService interface {
	CreateOrderWithSaga(
		ctx context.Context,
		req *dto.CreateOrderRequest,
	) (*dto.OrderResponse, error)
	CreateOrderWithTemporal(
		ctx context.Context,
		req *dto.CreateOrderRequest,
	) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	GetOrdersByCustomer(
		ctx context.Context,
		customerID uuid.UUID,
		req dto.GetOrdersRequest,
	) ([]dto.OrderResponse, *pkgdto.OffsetPagination, error)
	GetOrders(
		ctx context.Context,
		req dto.GetOrdersRequest,
	) ([]dto.OrderResponse, *pkgdto.OffsetPagination, error)
	UpdateOrderStatus(
		ctx context.Context,
		id uuid.UUID,
		status constant.OrderStatus,
	) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, req *dto.CancelOrderRequest) error
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

// orderService implements the OrderService.
type orderService struct {
	dataStore              repository.DataStore
	logger                 logger.Logger
	orderLifecycleProducer kafka.Producer
	sagaOrchestrator       saga.Orchestrator
	temporalClient         *client.TemporalClient
	config                 *config.Config
}

// NewOrderService creates a new instance of orderService.
func NewOrderService(
	cfg *config.Config,
	dataStore repository.DataStore,
	appLogger logger.Logger,
	orderLifecycleProducer kafka.Producer,
	sagaOrchestrator saga.Orchestrator,
	temporalClient *client.TemporalClient,
) OrderService {
	return &orderService{
		dataStore:              dataStore,
		logger:                 appLogger,
		orderLifecycleProducer: orderLifecycleProducer,
		sagaOrchestrator:       sagaOrchestrator,
		temporalClient:         temporalClient,
		config:                 cfg,
	}
}

// CreateOrderWithTemporal handles POST /orders/temporal with Temporal processing.
func (s *orderService) CreateOrderWithTemporal(
	ctx context.Context,
	req *dto.CreateOrderRequest,
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
		if err = lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	var res *dto.OrderResponse

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		existingOrder, errExist := orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if errExist != nil && errExist.Error() != constant.OrderNotFoundErrorMessage {
			return errExist
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

		newOrder, newErr := entity.NewOrder(req.CustomerID, req.IdempotencyKey, "IDR", orderItems)
		if newErr != nil {
			return fmt.Errorf("failed to create order entity: %w", newErr)
		}

		if err = newOrder.UpdateStatus(constant.OrderStatusPending); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		savedOrder, createErr := orderRepo.Create(ctx, newOrder)
		if createErr != nil {
			return createErr
		}

		// Extract user authentication info from context
		userAuth, userErr := echoutils.GetUserAuthContexts(ctx)
		if userErr != nil {
			return userErr
		}

		// Start Temporal workflow
		temporalReq := dto.TemporalOrderSagaRequest{
			Order:    savedOrder,
			Shipping: &req.Shipping,
			UserAuth: &userAuth,
		}

		workflowOptions := s.temporalClient.CreateOrderWorkflowOptions(savedOrder.ID)

		workflowRun, wfErr := s.temporalClient.Client.ExecuteWorkflow(
			ctx,
			workflowOptions,
			constant.OrderSagaWorkflowName,
			temporalReq,
			s.config.Temporal,
		)
		if wfErr != nil {
			s.logger.Errorf(
				"Failed to start Temporal workflow for order %s: %v",
				savedOrder.ID,
				wfErr,
			)

			return fmt.Errorf("failed to start order processing workflow: %w", wfErr)
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
func (s *orderService) GetOrder(
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
func (s *orderService) GetOrdersByCustomer(
	ctx context.Context,
	customerID uuid.UUID,
	req dto.GetOrdersRequest,
) ([]dto.OrderResponse, *pkgdto.OffsetPagination, error) {
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

	pagination := pageutils.NewOffsetPagination(total, req.Page, req.Limit)

	return res, pagination, nil
}

// GetOrders retrieves all orders with pagination.
func (s *orderService) GetOrders(
	ctx context.Context,
	req dto.GetOrdersRequest,
) ([]dto.OrderResponse, *pkgdto.OffsetPagination, error) {
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

	pagination := pageutils.NewOffsetPagination(total, req.Page, req.Limit)

	return res, pagination, nil
}

// UpdateOrderStatus updates only the status of an order.
func (s *orderService) UpdateOrderStatus(
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
		if err = existingOrder.UpdateStatus(status); err != nil {
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
		if err = s.orderLifecycleProducer.Send(ctx, evt); err != nil {
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
func (s *orderService) CancelOrder(
	ctx context.Context,
	req *dto.CancelOrderRequest,
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
		if err = lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check if order exists
		existingOrder, errExist := orderRepo.FindByID(ctx, req.OrderID)
		if errExist != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Check if order can be canceled
		if !existingOrder.CanBeCancelled() {
			return httperror.NewBadRequestError(
				fmt.Sprintf("order cannot be canceled in current status: %s", existingOrder.Status),
			)
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
		if err = existingOrder.UpdateStatus(constant.OrderStatusCanceled); err != nil {
			return httperror.NewBadRequestError("failed to cancel order entity")
		}

		// Save updated order
		updatedOrder, updateErr := orderRepo.Update(ctx, existingOrder)
		if updateErr != nil {
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
		if err = s.orderLifecycleProducer.Send(ctx, evt); err != nil {
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
func (s *orderService) RequestPaymentOrder(
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
		if err = lockRepo.Release(ctx, lock); err != nil {
			s.logger.Warnf("failed to release lock: %v", err)
		}
	}()

	res := new(dto.OrderResponse)

	err = s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()
		outboxRepo := ds.OutboxRepository()

		// Check if order exists
		existingOrder, findErr := orderRepo.FindByID(ctx, id)
		if findErr != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Check if order can be paid
		if !existingOrder.CanBePaid() {
			return httperror.NewBadRequestError(
				fmt.Sprintf("order cannot be paid in current status: %s", existingOrder.Status),
			)
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

		payload, marshalErr := json.Marshal(evt)
		if marshalErr != nil {
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

		if err = outboxRepo.Create(ctx, outboxEvent); err != nil {
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
func (s *orderService) NotifyOrderFailure(
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
