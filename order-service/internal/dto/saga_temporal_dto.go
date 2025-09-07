package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// CreateShippingResponse represents the result of creating shipping.
type CreateShippingResponse struct {
	ShippingID     uuid.UUID `json:"shipping_id"`
	TrackingNumber string    `json:"tracking_number"`
}

// SendOrderConfirmationRequest represents input for sending order confirmation.
type SendOrderConfirmationRequest struct {
	Order          *entity.Order    `json:"order"`
	Products       []entity.Product `json:"products"`
	TrackingNumber string           `json:"tracking_number"`
	CustomerEmail  string           `json:"customer_email"`
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
	ShippingCost     *decimal.Decimal     `json:"shipping_cost"`
	CustomerEmail    string               `json:"customer_email"`
	CompletedSteps   map[string]bool      `json:"completed_steps"`
}

// ReserveProductsAndCalculateResponse represents the result of reserving products and calculating order.
type ReserveProductsAndCalculateResponse struct {
	CalculatedOrder  *entity.Order    `json:"calculated_order"`
	ReservedProducts []entity.Product `json:"reserved_products"`
	CustomerEmail    string           `json:"customer_email"`
}

// ProcessFulfillmentResponse represents the result of processing fulfillment.
type ProcessFulfillmentResponse struct {
	ShippingID     uuid.UUID       `json:"shipping_id"`
	ShippingCost   decimal.Decimal `json:"shipping_cost"`
	TrackingNumber string          `json:"tracking_number"`
}

// SetFinalOrderPricesRequest represents input for setting final order prices.
type SetFinalOrderPricesRequest struct {
	Order        *entity.Order   `json:"order"`
	ShippingCost decimal.Decimal `json:"shipping_cost"`
}

// ConfirmProductsDeductionRequest represents input for confirming product deduction.
type ConfirmProductsDeductionRequest struct {
	Order            *entity.Order    `json:"order"`
	ReservedProducts []entity.Product `json:"reserved_products"`
}
