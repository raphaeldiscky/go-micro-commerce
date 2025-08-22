package dto

import "github.com/google/uuid"

// PayOrderRequest represents the request to pay an order.
type PayOrderRequest struct {
	OrderID       uuid.UUID `json:"order_id" validate:"required"`
	CustomerID    uuid.UUID `json:"customer_id" validate:"required"`
	CustomerEmail string    `json:"customer_email" validate:"required,email"`
	RequestID     uuid.UUID `json:"request_id" validate:"required"`
}

// PayOrderResponse represents the response for paying an order.
type PayOrderResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	OrderID uuid.UUID `json:"order_id"`
}
