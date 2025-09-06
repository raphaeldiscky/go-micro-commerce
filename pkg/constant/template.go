package constant

// TemplateIDType is a type for email template IDs.
type TemplateIDType string

// Email template IDs for different notification types.
const (
	TemplateOrderConfirmation   TemplateIDType = "order_confirmation"
	TemplateOrderShipped        TemplateIDType = "order_shipped"
	TemplateOrderCanceled       TemplateIDType = "order_canceled"
	TemplatePaymentConfirmation TemplateIDType = "payment_confirmation"
)
