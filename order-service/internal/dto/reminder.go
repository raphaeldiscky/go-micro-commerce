package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// PaymentReminderWorkflowRequest contains parameters specific to payment reminders.
type PaymentReminderWorkflowRequest struct {
	OrderID          uuid.UUID        `json:"order_id"`
	ReservedProducts []entity.Product `json:"reserved_products"`
	CustomerEmail    string           `json:"customer_email"`
	PaymentID        uuid.UUID        `json:"payment_id"`
	TotalPrice       decimal.Decimal  `json:"total_price"`
	Currency         string           `json:"currency"`
	TaskQueue        string           `json:"task_queue"`
}

// PaymentReminderRequest contains parameters for sending a payment reminder.
type PaymentReminderRequest struct {
	OrderID          uuid.UUID        `json:"order_id"`
	CustomerEmail    string           `json:"customer_email"`
	ReservedProducts []entity.Product `json:"reserved_products"`
	PaymentID        uuid.UUID        `json:"payment_id"`
	TotalPrice       decimal.Decimal  `json:"total_price"`
	Currency         string           `json:"currency"`
	ReminderCount    int              `json:"reminder_count"`
	MaxReminders     int              `json:"max_reminders"`
}
