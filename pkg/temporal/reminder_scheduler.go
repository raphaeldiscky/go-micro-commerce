// Package temporal provides utilities for scheduling Temporal workflows.
package temporal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
)

// ReminderScheduler provides specialized scheduling for reminder workflows with calendar.
type ReminderScheduler struct {
	manager ScheduleManager
}

// NewReminderScheduler creates a new ReminderScheduler.
func NewReminderScheduler(manager ScheduleManager) *ReminderScheduler {
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

	if len(req.Config.ExecutionTimes) == 0 {
		return nil, errors.New("reminder times are required")
	}

	// Calculate specific execution times based on intervals
	baseTime := req.Config.BaseTime
	if baseTime.IsZero() {
		baseTime = time.Now().UTC()
	}

	executionTimes := make([]time.Time, len(req.Config.ExecutionTimes))
	for i, time := range req.Config.ExecutionTimes {
		executionTimes[i] = baseTime.Add(time)
	}

	builder := NewScheduleConfigBuilder().
		WithID(req.ID).
		WithDescription(req.Description).
		WithCalendarSpec(executionTimes)

	// Set workflow action with task queue if provided
	workflowOptions := &client.StartWorkflowOptions{}
	if req.TaskQueue != "" {
		workflowOptions.TaskQueue = req.TaskQueue
	}

	builder = builder.WithWorkflowAction(req.WorkflowType, req.Input, workflowOptions)

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
