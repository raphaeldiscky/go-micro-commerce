package dto

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// CreateShippingResponse represents the result of creating shipping.
type CreateShippingResponse struct {
	ShippingID     uuid.UUID `json:"shipping_id"`
	TrackingNumber string    `json:"tracking_number"`
}

// SendOrderConfirmationRequest represents input for sending order confirmation.
type SendOrderConfirmationRequest struct {
	Order          *entity.Order `json:"order"`
	TrackingNumber string        `json:"tracking_number"`
}

// RefundPaymentGatewayRequest represents input for refunding payment.
type RefundPaymentGatewayRequest struct {
	Order     *entity.Order `json:"order"`
	PaymentID uuid.UUID     `json:"payment_id"`
}

// TemporalOrderSagaRequest represents the input for the order saga workflow.
type TemporalOrderSagaRequest struct {
	Order *entity.Order `json:"order"`
}

// TemporalOrderSagaResponse represents the result of the order saga workflow.
type TemporalOrderSagaResponse struct {
	Success        bool                 `json:"success"`
	OrderID        uuid.UUID            `json:"order_id"`
	PaymentID      *uuid.UUID           `json:"payment_id,omitempty"`
	ShippingID     *uuid.UUID           `json:"shipping_id,omitempty"`
	TrackingNumber *string              `json:"tracking_number,omitempty"`
	Pricing        *entity.OrderPricing `json:"pricing,omitempty"`
	Error          string               `json:"error,omitempty"`
}

// TemporalWorkflowState holds the state of the workflow execution.
type TemporalWorkflowState struct {
	OrderID          uuid.UUID            `json:"order_id"`
	ReservedProducts []entity.Product     `json:"reserved_products"`
	Pricing          *entity.OrderPricing `json:"pricing"`
	PaymentID        *uuid.UUID           `json:"payment_id"`
	ShippingID       *uuid.UUID           `json:"shipping_id"`
	TrackingNumber   *string              `json:"tracking_number"`
	CompletedSteps   map[string]bool      `json:"completed_steps"`
}
