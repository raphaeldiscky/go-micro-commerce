package constant

import "time"

const (
	// FirstReminderSequence represents the first reminder in the sequence.
	FirstReminderSequence = 1
	// SecondReminderSequence represents the second reminder in the sequence.
	SecondReminderSequence = 2
	// MaxPaymentReminders represents the max number of payment reminders.
	MaxPaymentReminders = 3
)

const (
	// PaymentReminderTimeout is the timeout for payment reminders.
	PaymentReminderTimeout = 5 * time.Minute
	// PaymentReminderInitialInterval is the initial interval for payment reminders.
	PaymentReminderInitialInterval = 1 * time.Second
	// PaymentReminderBackoffCoefficient is the backoff coefficient for payment reminders.
	PaymentReminderBackoffCoefficient = 2.0
	// PaymentReminderMaxInterval is the Max interval for payment reminders.
	PaymentReminderMaxInterval = 30 * time.Second
	// PaymentReminderMaxAttempts is the Max number of attempts for payment reminders.
	PaymentReminderMaxAttempts = 3
)

const (
	// SendPaymentReminderActivity is the activity name for sending payment reminders.
	SendPaymentReminderActivity = "SendPaymentReminderActivity"
	// CheckPaymentStatusActivity is the activity name for checking payment status.
	CheckPaymentStatusActivity = "CheckPaymentStatusActivity"
)

const (
	// PaymentReminderWorkflowType is the workflow type for payment reminders.
	PaymentReminderWorkflowType = "PaymentReminderWorkflow"
)

// GetPaymentReminderExecutionTimes returns the execution times for payment reminders.
func GetPaymentReminderExecutionTimes() []time.Duration {
	return []time.Duration{
		1 * time.Minute, // First reminder after 1 minute
		2 * time.Minute, // Second reminder after 2 minutes
	}
}
