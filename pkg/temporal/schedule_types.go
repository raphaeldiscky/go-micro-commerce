package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
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

// WorkflowConfig contains configuration for reminder schedules.
type WorkflowConfig struct {
	ExecutionTimes []time.Duration
	Timezone       *time.Location
	BaseTime       time.Time
}
