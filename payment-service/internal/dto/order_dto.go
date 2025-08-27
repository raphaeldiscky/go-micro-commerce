// Package dto contains data transfer objects for product service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/entity"
)

// CreateOrderItemRequest represents an item in create order request.
type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity"   validate:"required,min=1"`
}

// CreateOrderRequest represents the request to create a new order.
type CreateOrderRequest struct {
	CustomerID     uuid.UUID
	CustomerEmail  string
	IdempotencyKey uuid.UUID                `json:"idempotency_key" validate:"required"` // generated from client
	Items          []CreateOrderItemRequest `json:"items"           validate:"required,min=1,dive"`
}

// ClientCreateOrderRequest represents the request to create a new order from the client.
type ClientCreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

// UpdateOrderItemRequest represents an item in update order request.
type UpdateOrderItemRequest struct {
	ProductID uuid.UUID       `json:"product_id" validate:"required"`
	Quantity  int             `json:"quantity"   validate:"required,min=1"`
	Price     decimal.Decimal `json:"price"      validate:"required,decimal_gt"`
}

// UpdateOrderRequest represents the request to update an order.
type UpdateOrderRequest struct {
	ID     uuid.UUID                `json:"id"               validate:"required"`
	Status *constant.OrderStatus    `json:"status,omitempty"`
	Items  []UpdateOrderItemRequest `json:"items,omitempty"  validate:"omitempty,dive"`
}

// OrderItemResponse represents an order item in API responses.
type OrderItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Quantity  int             `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
}

// OrderResponse represents an order in API responses.
type OrderResponse struct {
	ID         uuid.UUID            `json:"id"`
	CustomerID uuid.UUID            `json:"customer_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Items      []OrderItemResponse  `json:"items"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

// GetOrdersRequest represents pagination and filtering parameters.
type GetOrdersRequest struct {
	Limit int64 `json:"limit" validate:"min=1,max=100"`
	Page  int64 `json:"page"  validate:"min=1"`
}

// UpdateOrderStatusRequest represents the request to update order status.
type UpdateOrderStatusRequest struct {
	Status  constant.OrderStatus `json:"status"   validate:"required"`
	OrderID uuid.UUID            `json:"order_id" validate:"required"`
}

// CancelOrderRequest represents the request to cancel an order.
type CancelOrderRequest struct {
	CustomerID     uuid.UUID
	CustomerEmail  string
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required"`
	Reason         string    `json:"reason"          validate:"required,min=5,max=255"`
}

// MapToOrderResponse converts domain entity to DTO response.
func MapToOrderResponse(order *entity.Order) *OrderResponse {
	return &OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      MapToOrderItemResponses(order.Items),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

// MapToOrderItemResponses converts domain entities to DTO responses.
func MapToOrderItemResponses(items []entity.OrderItem) []OrderItemResponse {
	var responses []OrderItemResponse
	for _, item := range items {
		responses = append(responses, OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return responses
}
