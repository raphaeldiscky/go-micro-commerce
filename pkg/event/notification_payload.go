package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NotificationRequestPayload represents a notification request event payload.
type NotificationRequestPayload struct {
	ID               uuid.UUID               `json:"id"`
	RecipientEmail   string                  `json:"recipient_email"`
	RecipientName    string                  `json:"recipient_name,omitempty"`
	NotificationType NotificationType        `json:"notification_type"`
	TemplateID       constant.TemplateIDType `json:"template_id"`
	Subject          string                  `json:"subject"`
	Priority         NotificationPriority    `json:"priority"`
	Data             map[string]any          `json:"data"`
	ScheduledFor     *time.Time              `json:"scheduled_for,omitempty"`
	ExpiresAt        *time.Time              `json:"expires_at,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
}

// NotificationType represents the type of notification.
type NotificationType string

const (
	// NotificationTypeEmail represents email notification type.
	NotificationTypeEmail NotificationType = "email"
	// NotificationTypeSMS represents SMS notification type.
	NotificationTypeSMS NotificationType = "sms"
	// NotificationTypePush represents push notification type.
	NotificationTypePush NotificationType = "push"
)

// NotificationPriority represents the priority of notification.
type NotificationPriority string

const (
	// NotificationPriorityLow represents low priority notification.
	NotificationPriorityLow NotificationPriority = "low"
	// NotificationPriorityNormal represents normal priority notification.
	NotificationPriorityNormal NotificationPriority = "normal"
	// NotificationPriorityHigh represents high priority notification.
	NotificationPriorityHigh NotificationPriority = "high"
)

//

// OrderConfirmationData represents data for order confirmation email.
type OrderConfirmationData struct {
	OrderID           uuid.UUID       `json:"order_id"`
	OrderNumber       string          `json:"order_number"`
	CustomerName      string          `json:"customer_name"`
	CustomerEmail     string          `json:"customer_email"`
	Items             []OrderItemData `json:"items"`
	Subtotal          decimal.Decimal `json:"subtotal"`
	ShippingCost      decimal.Decimal `json:"shipping_cost"`
	TotalTax          decimal.Decimal `json:"total_tax"`
	TotalDiscount     decimal.Decimal `json:"total_discount"`
	TotalPrice        decimal.Decimal `json:"total_price"`
	Currency          string          `json:"currency"`
	TrackingNumber    string          `json:"tracking_number,omitempty"`
	EstimatedDelivery *time.Time      `json:"estimated_delivery,omitempty"`
	OrderDate         time.Time       `json:"order_date"`
}

// OrderItemData represents an order item for email template.
type OrderItemData struct {
	ProductName   string          `json:"product_name"`
	Quantity      int64           `json:"quantity"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	TaxRate       decimal.Decimal `json:"tax_rate"`
	TotalTax      decimal.Decimal `json:"total_tax"`
	TotalDiscount decimal.Decimal `json:"total_discount"`
	TotalPrice    decimal.Decimal `json:"total_price"`
}
