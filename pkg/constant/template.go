package constant

// TemplateIDType is a type for email template IDs.
type TemplateIDType string

// Email template IDs for different notification types.
const (
	TemplateOrderConfirmed       TemplateIDType = "order_confirmed"
	TemplateOrderShipped         TemplateIDType = "order_shipped"
	TemplateOrderCanceled        TemplateIDType = "order_canceled"
	TemplateOrderDelivered       TemplateIDType = "order_delivered"
	TemplateOrderPaymentRequired TemplateIDType = "order_payment_required"
)
