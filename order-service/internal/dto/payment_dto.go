package dto

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PayOrderRequest represents the request to pay an order.
type PayOrderRequest struct {
	CustomerID     uuid.UUID
	CustomerEmail  string
	IdempotencyKey uuid.UUID               `json:"idempotency_key" validate:"required"`
	PaymentMethod  constant.PaymentMethod  `json:"payment_method"  validate:"required"`
	PaymentGateway constant.PaymentGateway `json:"payment_gateway" validate:"required"`
}

// PayOrderResponse represents the response for paying an order.
type PayOrderResponse struct {
	OrderID uuid.UUID            `json:"order_id"`
	Status  constant.OrderStatus `json:"status"`
}

// PaymentResponse represents the response from payment service.
type PaymentResponse struct {
	PaymentID uuid.UUID              `json:"payment_id"`
	Status    constant.PaymentStatus `json:"status"`
	OrderID   uuid.UUID              `json:"order_id"`
	Error     error                  `json:"error,omitempty"`
}
