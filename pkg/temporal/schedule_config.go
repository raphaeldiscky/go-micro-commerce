package temporal

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
)

// ScheduleConfigBuilder provides a fluent interface for building schedule configurations.
type ScheduleConfigBuilder struct {
	options ScheduleOptions
	errors  []error
}

// NewScheduleConfigBuilder creates a new ScheduleConfigBuilder.
func NewScheduleConfigBuilder() *ScheduleConfigBuilder {
	return &ScheduleConfigBuilder{
		options: ScheduleOptions{},
		errors:  make([]error, 0),
	}
}

// WithID sets the schedule ID.
func (b *ScheduleConfigBuilder) WithID(id string) *ScheduleConfigBuilder {
	if id == "" {
		b.errors = append(b.errors, errors.New("schedule ID cannot be empty"))
		return b
	}

	b.options.ID = id

	return b
}

// WithDescription sets the schedule description.
func (b *ScheduleConfigBuilder) WithDescription(description string) *ScheduleConfigBuilder {
	b.options.Description = description
	return b
}

// WithIntervalSpec sets an interval-based schedule specification.
func (b *ScheduleConfigBuilder) WithIntervalSpec(
	interval time.Duration,
	jitter time.Duration,
) *ScheduleConfigBuilder {
	if interval <= 0 {
		b.errors = append(b.errors, errors.New("interval must be positive"))
		return b
	}

	b.options.Spec = client.ScheduleSpec{
		Intervals: []client.ScheduleIntervalSpec{
			{
				Every:  interval,
				Offset: jitter,
			},
		},
	}

	return b
}

// WithCronSpec sets a cron-based schedule specification.
func (b *ScheduleConfigBuilder) WithCronSpec(
	cronExpr string,
	timezone string,
) *ScheduleConfigBuilder {
	if cronExpr == "" {
		b.errors = append(b.errors, errors.New("cron expression cannot be empty"))
		return b
	}

	if timezone != "" {
		_, err := time.LoadLocation(timezone)
		if err != nil {
			b.errors = append(b.errors, fmt.Errorf("invalid timezone: %w", err))
			return b
		}
	}

	b.options.Spec = client.ScheduleSpec{
		CronExpressions: []string{cronExpr},
	}

	return b
}

// WithCalendarSpec sets a calendar-based schedule specification for one-time or specific executions.
func (b *ScheduleConfigBuilder) WithCalendarSpec(
	executionTimes []time.Time,
) *ScheduleConfigBuilder {
	if len(executionTimes) == 0 {
		b.errors = append(b.errors, errors.New("at least one execution time is required"))
		return b
	}

	calendars := make([]client.ScheduleCalendarSpec, len(executionTimes))
	for i, execTime := range executionTimes {
		calendars[i] = client.ScheduleCalendarSpec{
			Year: []client.ScheduleRange{{Start: execTime.Year(), End: execTime.Year()}},
			Month: []client.ScheduleRange{
				{Start: int(execTime.Month()), End: int(execTime.Month())},
			},
			DayOfMonth: []client.ScheduleRange{{Start: execTime.Day(), End: execTime.Day()}},
			Hour:       []client.ScheduleRange{{Start: execTime.Hour(), End: execTime.Hour()}},
			Minute:     []client.ScheduleRange{{Start: execTime.Minute(), End: execTime.Minute()}},
			Second:     []client.ScheduleRange{{Start: execTime.Second(), End: execTime.Second()}},
		}
	}

	b.options.Spec = client.ScheduleSpec{
		Calendars: calendars,
	}

	return b
}

// WithWorkflowAction sets a workflow action for the schedule.
func (b *ScheduleConfigBuilder) WithWorkflowAction(
	workflowType string,
	input interface{},
	options *client.StartWorkflowOptions,
) *ScheduleConfigBuilder {
	if workflowType == "" {
		b.errors = append(b.errors, errors.New("workflow type cannot be empty"))
		return b
	}

	action := &client.ScheduleWorkflowAction{
		Workflow: workflowType,
		Args:     []interface{}{input},
	}

	if options != nil {
		action.TaskQueue = options.TaskQueue
	}

	b.options.Action = action

	return b
}

// Build builds the schedule options and validates the configuration.
func (b *ScheduleConfigBuilder) Build() (ScheduleOptions, error) {
	if b.options.ID == "" {
		b.errors = append(b.errors, errors.New("schedule ID is required"))
	}

	if b.options.Action == nil {
		b.errors = append(b.errors, errors.New("schedule action is required"))
	}

	if len(b.options.Spec.Intervals) == 0 && len(b.options.Spec.CronExpressions) == 0 &&
		len(b.options.Spec.Calendars) == 0 {
		b.errors = append(
			b.errors,
			errors.New("schedule specification is required (interval, cron, or calendar)"),
		)
	}

	if len(b.errors) > 0 {
		return ScheduleOptions{}, fmt.Errorf("validation errors: %v", b.errors)
	}

	return b.options, nil
}

// HourlySchedule creates a schedule that runs every hour.
func HourlySchedule(id, workflowType string, input interface{}) *ScheduleConfigBuilder {
	return NewScheduleConfigBuilder().
		WithID(id).
		WithDescription("Hourly schedule").
		WithIntervalSpec(time.Hour, 0).
		WithWorkflowAction(workflowType, input, nil)
}

// DailySchedule creates a schedule that runs daily at a specific hour.
func DailySchedule(id, workflowType string, input interface{}, hour int) *ScheduleConfigBuilder {
	cronExpr := fmt.Sprintf("0 %d * * *", hour)

	return NewScheduleConfigBuilder().
		WithID(id).
		WithDescription("Daily schedule").
		WithCronSpec(cronExpr, "UTC").
		WithWorkflowAction(workflowType, input, nil)
}

// WeeklySchedule creates a schedule that runs weekly on a specific day and hour.
func WeeklySchedule(
	id, workflowType string,
	input interface{},
	dayOfWeek, hour int,
) *ScheduleConfigBuilder {
	cronExpr := fmt.Sprintf("0 %d * * %d", hour, dayOfWeek)

	return NewScheduleConfigBuilder().
		WithID(id).
		WithDescription("Weekly schedule").
		WithCronSpec(cronExpr, "UTC").
		WithWorkflowAction(workflowType, input, nil)
}

// IntervalSchedule creates a schedule with custom intervals.
func IntervalSchedule(
	id, workflowType string,
	input interface{},
	interval time.Duration,
) *ScheduleConfigBuilder {
	return NewScheduleConfigBuilder().
		WithID(id).
		WithDescription("Reminder schedule").
		WithIntervalSpec(interval, 0).
		WithWorkflowAction(workflowType, input, nil)
}
