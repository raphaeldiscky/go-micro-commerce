package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
)

// Shedule

// ReminderType represents different types of reminders.
type ReminderType string

const (
	// ReminderTypePayment represents payment reminders.
	ReminderTypePayment ReminderType = "payment"
	// ReminderTypeSubscription represents subscription reminders.
	ReminderTypeSubscription ReminderType = "subscription"
	// ReminderTypeCart represents abandoned cart reminders.
	ReminderTypeCart ReminderType = "cart"
	// ReminderTypePromotion represents promotional reminders.
	ReminderTypePromotion ReminderType = "promotion"
)

// ScheduleOptions contains options for creating a schedule.
type ScheduleOptions struct {
	ID          string
	Description string
	Spec        client.ScheduleSpec
	Action      client.ScheduleAction
	StartAt     *time.Time
	EndAt       *time.Time
}

// ReminderConfig contains configuration for reminder schedules.
type ReminderConfig struct {
	Type           ReminderType
	ExecutionTimes []time.Duration
	Timezone       *time.Location
	BaseTime       time.Time
}
