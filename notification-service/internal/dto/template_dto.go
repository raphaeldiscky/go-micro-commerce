// Package dto provides data transfer objects for the notification service.
package dto

// OrderConfirmationTemplateData represents data for order confirmation email template.
type OrderConfirmationTemplateData struct {
	CustomerName  string                    `json:"customer_name"`
	OrderNumber   string                    `json:"order_number"`
	OrderID       string                    `json:"order_id"`
	OrderDate     string                    `json:"order_date"`
	CustomerEmail string                    `json:"customer_email"`
	Items         []OrderItemTemplateData   `json:"items"`
	Subtotal      string                    `json:"subtotal"`
	ShippingCost  string                    `json:"shipping_cost"`
	TotalTax      string                    `json:"total_tax"`
	TotalDiscount string                    `json:"total_discount"`
	TotalPrice    string                    `json:"total_price"`
	Currency      string                    `json:"currency"`
	Shipping      *ShippingInfoTemplateData `json:"shipping,omitempty"`
}

// OrderItemTemplateData represents an order item for email template.
type OrderItemTemplateData struct {
	ProductName string `json:"product_name"`
	Quantity    int64  `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	TotalPrice  string `json:"total_price"`
	Currency    string `json:"currency"`
}

// ShippingInfoTemplateData represents shipping information for email template.
type ShippingInfoTemplateData struct {
	TrackingNumber    string `json:"tracking_number"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
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

// PaymentConfirmationTemplateData represents data for payment confirmation email template.
type PaymentConfirmationTemplateData struct {
	RecipientName string `json:"recipient_name"`
	Amount        string `json:"amount,omitempty"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

// GenericTemplateData represents data for generic email template.
type GenericTemplateData struct {
	Subject       string                 `json:"subject"`
	RecipientName string                 `json:"recipient_name"`
	Data          map[string]interface{} `json:"data,omitempty"`
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
