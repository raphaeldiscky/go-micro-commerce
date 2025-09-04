// Package dto contains data transfer objects for product service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// CreateOrderItemRequest represents an item in create order request.
type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity"   validate:"required,min=1"`
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

// OrderItemResponse represents an order item in API responses.
type OrderItemResponse struct {
	ID            uuid.UUID       `json:"id"`
	ProductID     uuid.UUID       `json:"product_id"`
	Quantity      int64           `json:"quantity"`
	Currency      string          `json:"currency"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	TotalPrice    decimal.Decimal `json:"total_price"`
	TotalTax      decimal.Decimal `json:"total_tax"`
	TotalDiscount decimal.Decimal `json:"total_discount"`
}

// OrderResponse represents an order in API responses.
type OrderResponse struct {
	ID            uuid.UUID            `json:"id"`
	CustomerID    uuid.UUID            `json:"customer_id"`
	Status        constant.OrderStatus `json:"status"`
	Currency      string               `json:"currency"`
	TotalPrice    decimal.Decimal      `json:"total_price"`
	TotalTax      decimal.Decimal      `json:"total_tax"`
	TotalDiscount decimal.Decimal      `json:"total_discount"`
	Items         []OrderItemResponse  `json:"items"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
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
