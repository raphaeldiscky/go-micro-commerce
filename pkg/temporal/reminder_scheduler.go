// Package temporal provides utilities for scheduling Temporal workflows.
package temporal

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/client"
)

// ReminderScheduler provides specialized scheduling for reminder workflows.
type ReminderScheduler struct {
	manager ScheduleManagerInterface
}

// NewReminderScheduler creates a new ReminderScheduler.
func NewReminderScheduler(manager ScheduleManagerInterface) *ReminderScheduler {
	return &ReminderScheduler{
		manager: manager,
	}
}

// ReminderScheduleRequest contains parameters for creating a reminder schedule.
type ReminderScheduleRequest struct {
	ID           string
	WorkflowType string
	Input        any
	Config       ReminderConfig
	TaskQueue    string
	Description  string
}

// CreateReminderSchedule creates a reminder schedule with the specified configuration.
func (rs *ReminderScheduler) CreateReminderSchedule(
	ctx context.Context,
	req ReminderScheduleRequest,
) (client.ScheduleHandle, error) {
	if req.ID == "" {
		return nil, errors.New("schedule ID is required")
	}

	if req.WorkflowType == "" {
		return nil, errors.New("workflow type is required")
	}

	if len(req.Config.Intervals) == 0 {
		return nil, errors.New("reminder intervals are required")
	}

	// Use the first interval for the schedule - the workflow itself will handle escalation
	interval := req.Config.Intervals[0]
	if interval <= 0 {
		return nil, errors.New("reminder interval must be positive")
	}

	builder := NewScheduleConfigBuilder().
		WithID(req.ID).
		WithDescription(req.Description).
		WithIntervalSpec(interval, 0)

	// Set workflow action with task queue if provided
	workflowOptions := &client.StartWorkflowOptions{}
	if req.TaskQueue != "" {
		workflowOptions.TaskQueue = req.TaskQueue
	}

	builder = builder.WithWorkflowAction(req.WorkflowType, req.Input, workflowOptions)

	// Add memo with reminder metadata
	memo := map[string]any{
		"reminderType": string(req.Config.Type),
		"maxReminders": req.Config.MaxReminders,
	}

	if req.Config.Timezone != nil {
		memo["timezone"] = req.Config.Timezone.String()
	}

	builder = builder.WithMemo(memo)

	// Add search attributes for easy filtering
	searchAttributes := map[string]any{
		"ReminderType": string(req.Config.Type),
		"EntityID":     req.ID,
	}
	builder = builder.WithSearchAttributes(searchAttributes)

	options, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build schedule configuration: %w", err)
	}

	return rs.manager.Create(ctx, options)
}

// CancelReminderSchedule cancels an active reminder schedule.
func (rs *ReminderScheduler) CancelReminderSchedule(ctx context.Context, scheduleID string) error {
	return rs.manager.Delete(ctx, scheduleID)
}

// PauseReminderSchedule pauses a reminder schedule.
func (rs *ReminderScheduler) PauseReminderSchedule(
	ctx context.Context,
	scheduleID, reason string,
) error {
	return rs.manager.Pause(ctx, scheduleID, reason)
}

// ResumeReminderSchedule resumes a paused reminder schedule.
func (rs *ReminderScheduler) ResumeReminderSchedule(
	ctx context.Context,
	scheduleID, reason string,
) error {
	return rs.manager.Resume(ctx, scheduleID, reason)
}

// GetReminderScheduleInfo retrieves information about a reminder schedule.
func (rs *ReminderScheduler) GetReminderScheduleInfo(
	ctx context.Context,
	scheduleID string,
) (*client.ScheduleDescription, error) {
	return rs.manager.Describe(ctx, scheduleID)
}

// ListReminderSchedules lists reminder schedules by type.
func (rs *ReminderScheduler) ListReminderSchedules(
	ctx context.Context,
	reminderType ReminderType,
) (client.ScheduleListIterator, error) {
	query := fmt.Sprintf("ReminderType = '%s'", string(reminderType))
	return rs.manager.List(ctx, query)
}
