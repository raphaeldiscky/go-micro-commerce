package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	pkgtemporal "github.com/raphaeldiscky/go-micro-commerce/pkg/temporal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/sagautils"
)

// PaymentReminderService handles payment reminder scheduling for orders.
type PaymentReminderService struct {
	reminderScheduler *pkgtemporal.ReminderScheduler
}

// NewPaymentReminderService creates a new PaymentReminderService.
func NewPaymentReminderService(
	reminderScheduler *pkgtemporal.ReminderScheduler,
) *PaymentReminderService {
	return &PaymentReminderService{
		reminderScheduler: reminderScheduler,
	}
}

// CreatePaymentReminderSchedule creates a payment reminder schedule for an order.
func (prs *PaymentReminderService) CreatePaymentReminderSchedule(
	ctx context.Context,
	req dto.PaymentReminderWorkflowRequest,
) (client.ScheduleHandle, error) {
	scheduleID := sagautils.CreatePaymentReminderID(req.OrderID)

	reminderRequest := pkgtemporal.ReminderScheduleRequest{
		ID:           scheduleID,
		WorkflowType: constant.PaymentReminderWorkflowType,
		Input:        req,
		Config: pkgtemporal.ReminderConfig{
			Type:           pkgtemporal.ReminderTypePayment,
			ExecutionTimes: constant.GetPaymentReminderExecutionTimes(),
			Timezone:       time.UTC,
		},
		TaskQueue:   req.TaskQueue,
		Description: fmt.Sprintf("Payment reminder for order %s", req.OrderID),
	}

	return prs.reminderScheduler.CreateReminderSchedule(ctx, reminderRequest)
}

// CancelPaymentReminderSchedule cancels a payment reminder schedule for an order.
func (prs *PaymentReminderService) CancelPaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.CancelReminderSchedule(ctx, scheduleID)
}

// PausePaymentReminderSchedule pauses a payment reminder schedule for an order.
func (prs *PaymentReminderService) PausePaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
	reason string,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.PauseReminderSchedule(ctx, scheduleID, reason)
}

// ResumePaymentReminderSchedule resumes a payment reminder schedule for an order.
func (prs *PaymentReminderService) ResumePaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
	reason string,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.ResumeReminderSchedule(ctx, scheduleID, reason)
}
