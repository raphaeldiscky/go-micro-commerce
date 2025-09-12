// Package dto provides data transfer objects for the notification service.
package dto

// OrderConfirmedTemplateData represents data for order confirmation email template.
type OrderConfirmedTemplateData struct {
	CustomerName      string                  `json:"customer_name"`
	CustomerEmail     string                  `json:"customer_email"`
	OrderID           string                  `json:"order_id"`
	OrderDate         string                  `json:"order_date"`
	Items             []OrderItemTemplateData `json:"items"`
	Subtotal          string                  `json:"subtotal"`
	ShippingCost      string                  `json:"shipping_cost"`
	TotalTax          string                  `json:"total_tax"`
	TotalDiscount     string                  `json:"total_discount"`
	TotalPrice        string                  `json:"total_price"`
	Currency          string                  `json:"currency"`
	TrackingNumber    *string                 `json:"tracking_number,omitempty"`
	EstimatedDelivery string                  `json:"estimated_delivery,omitempty"`
}

// OrderDeliveredTemplateData represents data for order delivered email template.
type OrderDeliveredTemplateData struct {
	CustomerName      string                  `json:"customer_name"`
	CustomerEmail     string                  `json:"customer_email"`
	OrderID           string                  `json:"order_id"`
	OrderDate         string                  `json:"order_date"`
	Items             []OrderItemTemplateData `json:"items"`
	Subtotal          string                  `json:"subtotal"`
	ShippingCost      string                  `json:"shipping_cost"`
	TotalTax          string                  `json:"total_tax"`
	TotalDiscount     string                  `json:"total_discount"`
	TotalPrice        string                  `json:"total_price"`
	Currency          string                  `json:"currency"`
	TrackingNumber    *string                 `json:"tracking_number,omitempty"`
	EstimatedDelivery string                  `json:"estimated_delivery,omitempty"`
	ActualDeliveryAt  string                  `json:"actual_delivery_at"`
}

// OrderPaymentRequiredTemplateData represents data for order waiting payment email template.
type OrderPaymentRequiredTemplateData struct {
	CustomerName    string                  `json:"customer_name"`
	CustomerEmail   string                  `json:"customer_email"`
	OrderID         string                  `json:"order_id"`
	OrderDate       string                  `json:"order_date"`
	Items           []OrderItemTemplateData `json:"items"`
	Subtotal        string                  `json:"subtotal"`
	ShippingCost    string                  `json:"shipping_cost"`
	TotalTax        string                  `json:"total_tax"`
	TotalDiscount   string                  `json:"total_discount"`
	TotalPrice      string                  `json:"total_price"`
	Currency        string                  `json:"currency"`
	PaymentDeadline string                  `json:"payment_deadline"`
	PaymentURL      string                  `json:"payment_url,omitempty"`
}

// OrderItemTemplateData represents an order item for email template.
type OrderItemTemplateData struct {
	ProductName string `json:"product_name"`
	Quantity    int64  `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	TotalPrice  string `json:"total_price"`
	Currency    string `json:"currency"`
}

// OrderShippedTemplateData represents data for order shipped email template.
type OrderShippedTemplateData struct {
	RecipientName  string `json:"recipient_name"`
	OrderNumber    string `json:"order_number,omitempty"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}

// OrderCanceledTemplateData represents data for order canceled email template.
type OrderCanceledTemplateData struct {
	RecipientName string `json:"recipient_name"`
	OrderNumber   string `json:"order_number,omitempty"`
}

// PaymentConfirmedTemplateData represents data for payment confirmation email template.
type PaymentConfirmedTemplateData struct {
	RecipientName string `json:"recipient_name"`
	Amount        string `json:"amount,omitempty"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

// GenericTemplateData represents data for generic email template.
type GenericTemplateData struct {
	Subject       string         `json:"subject"`
	RecipientName string         `json:"recipient_name"`
	Data          map[string]any `json:"data,omitempty"`
}

// EmailVerificationTemplateData represents data for email verification template.
type EmailVerificationTemplateData struct {
	RecipientName   string `json:"recipient_name"`
	VerificationURL string `json:"verification_url"`
	TokenExpiresAt  string `json:"token_expires_at"`
}

// UserVerifiedTemplateData represents data for user verified email template.
type UserVerifiedTemplateData struct {
	RecipientName string `json:"recipient_name"`
}
