// Package dto provides data transfer objects for the payment service.
package dto

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/entity"
)

// PaymentRequest represents the request to pay a payment.
type PaymentRequest struct {
	CustomerID     uuid.UUID
	CustomerEmail  string
	IdempotencyKey uuid.UUID              `json:"idempotency_key" validate:"required"`
	PaymentMethod  constant.PaymentMethod `json:"payment_method"  validate:"required"`
}

// PaymentResponse represents the response for paying a payment.
type PaymentResponse struct {
	OrderID uuid.UUID              `json:"order_id"`
	Status  constant.PaymentStatus `json:"status"`
}

// MapToPaymentResponse converts domain entity to DTO response.
func MapToPaymentResponse(payment *entity.Payment) *PaymentResponse {
	return &PaymentResponse{
		OrderID: payment.ID,
		Status:  payment.Status,
	}
}
