package temporal

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

// PaymentReminderScheduler handles payment reminder scheduling for orders.
type PaymentReminderScheduler struct {
	reminderScheduler *pkgtemporal.ReminderScheduler
}

// NewPaymentReminderScheduler creates a new PaymentReminderScheduler.
func NewPaymentReminderScheduler(
	reminderScheduler *pkgtemporal.ReminderScheduler,
) *PaymentReminderScheduler {
	return &PaymentReminderScheduler{
		reminderScheduler: reminderScheduler,
	}
}

// CreatePaymentReminderSchedule creates a payment reminder schedule for an order.
func (prs *PaymentReminderScheduler) CreatePaymentReminderSchedule(
	ctx context.Context,
	req dto.PaymentReminderWorkflowRequest,
) (client.ScheduleHandle, error) {
	reminderWorkflowRequest := pkgtemporal.ReminderScheduleRequest{
		ID:           sagautils.CreatePaymentReminderID(req.OrderID),
		WorkflowType: constant.PaymentReminderWorkflowType,
		Input:        req,
		Config: pkgtemporal.ReminderConfig{
			Type:           pkgtemporal.ReminderTypePayment,
			ExecutionTimes: constant.GetPaymentReminderWorkflowExecutionTimes(),
			Timezone:       time.UTC,
		},
		TaskQueue:   req.TaskQueue,
		Description: fmt.Sprintf("Payment reminder for order %s", req.OrderID),
		StartAt:     time.Now().UTC().Add(0),
		// https://community.temporal.io/t/what-exactly-happens-when-the-endat-time-is-reached-for-a-schedule/16747
		EndAt: time.Now().UTC().Add(1 * time.Minute), // will deleted after 1 week if no future task
	}

	return prs.reminderScheduler.CreateReminderSchedule(ctx, reminderWorkflowRequest)
}

// CancelPaymentReminderSchedule cancels a payment reminder schedule for an order.
func (prs *PaymentReminderScheduler) CancelPaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.CancelReminderSchedule(ctx, scheduleID)
}

// PausePaymentReminderSchedule pauses a payment reminder schedule for an order.
func (prs *PaymentReminderScheduler) PausePaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
	reason string,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.PauseReminderSchedule(ctx, scheduleID, reason)
}

// ResumePaymentReminderSchedule resumes a payment reminder schedule for an order.
func (prs *PaymentReminderScheduler) ResumePaymentReminderSchedule(
	ctx context.Context,
	orderID uuid.UUID,
	reason string,
) error {
	scheduleID := sagautils.CreatePaymentReminderID(orderID)
	return prs.reminderScheduler.ResumeReminderSchedule(ctx, scheduleID, reason)
}
