package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
)

// ScheduleType represents different types of schedules.
type ScheduleType string

const (
	// ScheduleTypeInterval represents interval-based schedules.
	ScheduleTypeInterval ScheduleType = "interval"
	// ScheduleTypeCron represents cron-based schedules.
	ScheduleTypeCron ScheduleType = "cron"
	// ScheduleTypeCalendar represents calendar-based schedules.
	ScheduleTypeCalendar ScheduleType = "calendar"
)

// ScheduleState represents the state of a schedule.
type ScheduleState string

const (
	// ScheduleStateActive represents an active schedule.
	ScheduleStateActive ScheduleState = "active"
	// ScheduleStatePaused represents a paused schedule.
	ScheduleStatePaused ScheduleState = "paused"
	// ScheduleStateCompleted represents a completed schedule.
	ScheduleStateCompleted ScheduleState = "completed"
	// ScheduleStateCancelled represents a cancelled schedule.
	ScheduleStateCancelled ScheduleState = "cancelled"
)

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

// ScheduleInfo contains metadata about a schedule.
type ScheduleInfo struct {
	ID          string
	State       ScheduleState
	Type        ScheduleType
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	NextRunTime *time.Time
	LastRunTime *time.Time
}

// ScheduleOptions contains options for creating a schedule.
type ScheduleOptions struct {
	ID               string
	Description      string
	Spec             client.ScheduleSpec
	Action           client.ScheduleAction
	Memo             map[string]any
	SearchAttributes map[string]any
}

// WorkflowScheduleAction represents a workflow action for a schedule.
type WorkflowScheduleAction struct {
	WorkflowType string
	Input        any
	Options      *client.StartWorkflowOptions
}

// ReminderWorkflowInput represents the input for reminder workflows.
type ReminderWorkflowInput interface {
	// GetWorkflowType returns the workflow type for this input
	GetWorkflowType() string
}

// ReminderMemo contains metadata for reminder schedules.
type ReminderMemo struct {
	ReminderType string `json:"reminderType"`
	MaxReminders int    `json:"maxReminders"`
	Timezone     string `json:"timezone,omitempty"`
}

// ReminderSearchAttributes contains search attributes for reminder schedules.
type ReminderSearchAttributes struct {
	ReminderType string `json:"reminderType"`
	EntityID     string `json:"entityID"`
}

// ReminderConfig contains configuration for reminder schedules.
type ReminderConfig struct {
	Type         ReminderType
	MaxReminders int
	Intervals    []time.Duration
	Timezone     *time.Location
}
