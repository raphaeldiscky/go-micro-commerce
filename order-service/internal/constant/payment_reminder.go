package constant

import "time"

const (
	// FirstReminderSequence represents the first reminder in the sequence.
	FirstReminderSequence = 1
	// SecondReminderSequence represents the second reminder in the sequence.
	SecondReminderSequence = 2
)

const (
	// PaymentReminderTimeout is the timeout for payment reminders.
	PaymentReminderTimeout = 5 * time.Minute
	// PaymentReminderInitialInterval is the initial interval for payment reminders.
	PaymentReminderInitialInterval = 1 * time.Second
	// PaymentReminderBackoffCoefficient is the backoff coefficient for payment reminders.
	PaymentReminderBackoffCoefficient = 2.0
	// PaymentReminderMaxInterval is the Max interval for payment reminders.
	PaymentReminderMaxInterval = 5 * time.Second
	// PaymentReminderMaxAttempts is the Max number of attempts for payment reminders.
	PaymentReminderMaxAttempts = 0
)

const (
	// SendPaymentReminderActivity is the activity name for sending payment reminders.
	SendPaymentReminderActivity WorkflowStep = "SendPaymentReminderActivity"
	// CheckPaymentStatusActivity is the activity name for checking payment status.
	CheckPaymentStatusActivity WorkflowStep = "CheckPaymentStatusActivity"
	// ExpireOrderPaymentActivity is the activity name for expiring order payment.
	ExpireOrderPaymentActivity WorkflowStep = "ExpireOrderPaymentActivity"
	// CancelPaymentReminderScheduleActivity is the activity name for canceling payment reminder schedule.
	CancelPaymentReminderScheduleActivity WorkflowStep = "CancelPaymentReminderScheduleActivity"
)

const (
	// PaymentReminderWorkflowType is the workflow type for payment reminders.
	PaymentReminderWorkflowType = "PaymentReminderWorkflow"
)

const (
	// FirstPaymentReminderDelay is the delay for first task scheduling.
	FirstPaymentReminderDelay = 10 * time.Second
	// SecondPaymentReminderDelay is the delay for second task scheduling.
	SecondPaymentReminderDelay = 20 * time.Second
	// ExpireOrderReminderDelay is the delay for cancel task scheduling.
	ExpireOrderReminderDelay = 30 * time.Second
)

// GetPaymentReminderWorkflowExecutionTimes returns the execution times for payment reminders.
func GetPaymentReminderWorkflowExecutionTimes() []time.Duration {
	return []time.Duration{
		FirstPaymentReminderDelay,
	}
}

const (
	// FirstPaymentReminderEmailSubject is the subject for the first payment reminder.
	FirstPaymentReminderEmailSubject = "Payment Reminder - Your Order is Waiting"
	// SecondPaymentReminderEmailSubject is the subject for the second payment reminder.
	SecondPaymentReminderEmailSubject = "Final Payment Reminder - Your Order Expires Soon"
	// OrderPaymentExpiredEmailSubject is the subject for the order cancellation email.
	OrderPaymentExpiredEmailSubject = "Your Order Has Expired - Payment Timeout"
)
