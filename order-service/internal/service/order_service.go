// Package service provides business logic for order operations.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/pageutils"

	pkgDto "github.com/raphaeldiscky/go-micro-template/pkg/dto"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/event"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/repository"
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
	UpdateOrder(ctx context.Context, req dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	UpdateOrderStatus(
		ctx context.Context,
		id uuid.UUID,
		status constant.OrderStatus,
	) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, id uuid.UUID) error
	PayOrder(ctx context.Context, req dto.PayOrderRequest) (*dto.OrderResponse, error)
}

// OrderService implements the OrderServiceInterface.
type OrderService struct {
	dataStore              repository.DataStore
	orderLifecycleProducer mq.KafkaProducerInterface
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(
	dataStore repository.DataStore,
	orderLifecycleProducer mq.KafkaProducerInterface,
) OrderServiceInterface {
	return &OrderService{
		dataStore:              dataStore,
		orderLifecycleProducer: orderLifecycleProducer,
	}
}

// CreateOrder creates a new order.
func (s *OrderService) CreateOrder(
	ctx context.Context,
	req dto.CreateOrderRequest,
) (*dto.OrderResponse, error) {
	res := new(dto.OrderResponse)

	err := s.dataStore.Atomic(ctx, func(tx repository.DataStore) error {
		orderRepo := tx.OrderRepository()

		// Convert DTO items to entity items
		var orderItems []entity.OrderItem
		for _, item := range req.Items {
			orderItem := entity.OrderItem{
				ID:        uuid.New(),
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
			orderItems = append(orderItems, orderItem)
		}

		// Create domain entity
		order, err := entity.NewOrder(req.CustomerID, orderItems)
		if err != nil {
			return httperror.NewInvalidRequestBodyError()
		}

		// Save to repository
		savedOrder, err := orderRepo.Create(ctx, order)
		if err != nil {
			return httperror.NewInternalServerError("failed to create order")
		}

		// Publish domain event
		evt := event.NewOrderLifecycleEvent(
			savedOrder.ID,
			constant.OrderStatusPending,
			savedOrder.CustomerID,
			savedOrder.TotalPrice,
		)

		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order created event")
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

// UpdateOrder updates an existing order.
func (s *OrderService) UpdateOrder(
	ctx context.Context,
	req dto.UpdateOrderRequest,
) (*dto.OrderResponse, error) {
	res := new(dto.OrderResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check if order exists
		existingOrder, err := orderRepo.FindByID(ctx, req.ID)
		if err != nil {
			return httperror.NewInternalServerError("failed to get order")
		}

		if existingOrder == nil {
			return httperror.NewOrderNotFoundError()
		}

		// Update status if provided
		if req.Status != nil {
			if err := existingOrder.UpdateStatus(*req.Status); err != nil {
				return httperror.NewBadRequestError("invalid order status")
			}
		}

		// Update items if provided
		if req.Items != nil {
			// Convert DTO items to entity items
			var orderItems []entity.OrderItem
			for _, item := range req.Items {
				orderItem := entity.OrderItem{
					ID:        uuid.New(),
					OrderID:   existingOrder.ID,
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
					Price:     item.Price,
				}
				orderItems = append(orderItems, orderItem)
			}

			// Update items in the order (this would need to be implemented in the entity)
			if err := existingOrder.UpdateItems(orderItems); err != nil {
				return httperror.NewBadRequestError("invalid order items")
			}
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, existingOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to update order")
		}

		// Publish domain event
		evt := event.NewOrderUpdatedEvent(
			updatedOrder.ID,
			updatedOrder.CustomerID,
			updatedOrder.Status,
			updatedOrder.TotalPrice,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order updated event")
		}

		res = dto.MapToOrderResponse(updatedOrder)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateOrderStatus updates only the status of an order.
func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	id uuid.UUID,
	status entity.OrderStatus,
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
		evt := event.NewOrderStatusUpdatedEvent(
			updatedOrder.ID,
			updatedOrder.CustomerID,
			updatedOrder.Status,
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
func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
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

		// Check if order can be cancelled
		if !existingOrder.CanBeCancelled() {
			return httperror.NewBadRequestError("order cannot be cancelled in current status")
		}

		// Update status to cancelled
		if err := existingOrder.UpdateStatus(entity.OrderStatusCanceled); err != nil {
			return httperror.NewBadRequestError("failed to cancel order")
		}

		// Save updated order
		if _, err := orderRepo.Update(ctx, existingOrder); err != nil {
			return httperror.NewInternalServerError("failed to cancel order")
		}

		// Publish domain event
		evt := event.NewOrderCancelledEvent(
			existingOrder.ID,
			existingOrder.CustomerID,
			existingOrder.TotalPrice,
		)
		if err := s.orderLifecycleProducer.Send(ctx, evt); err != nil {
			return httperror.NewInternalServerError("failed to send order cancelled event")
		}

		return nil
	})

	return err
}

// PayOrder processes payment for an order.
func (s *OrderService) PayOrder(
	ctx context.Context,
	req dto.PayOrderRequest,
) (*dto.OrderResponse, error) {
	res := new(dto.OrderResponse)

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		orderRepo := ds.OrderRepository()

		// Check if order exists
		existingOrder, err := orderRepo.FindByID(ctx, req.ID)
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

		// Update status to paid
		if err := existingOrder.UpdateStatus(entity.OrderStatusPaid); err != nil {
			return httperror.NewBadRequestError("failed to pay order")
		}

		// Save updated order
		updatedOrder, err := orderRepo.Update(ctx, existingOrder)
		if err != nil {
			return httperror.NewInternalServerError("failed to pay order")
		}

		// Publish domain event
		evt := event.NewOrderPaidEvent(
			updatedOrder.ID,
			updatedOrder.CustomerID,
			updatedOrder.TotalPrice,
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
