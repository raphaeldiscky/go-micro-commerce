// Package temporal provides utilities for workflows with temporal Schedules.
package temporal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// WorkflowScheduler provides specialized scheduling for Workflow workflows with calendar.
type WorkflowScheduler struct {
	manager ScheduleManager
}

// NewWorkflowScheduler creates a new WorkflowScheduler.
func NewWorkflowScheduler(manager ScheduleManager) *WorkflowScheduler {
	return &WorkflowScheduler{
		manager: manager,
	}
}

// WorkflowScheduleRequest contains parameters for creating a Workflow schedule.
type WorkflowScheduleRequest struct {
	ID           string
	WorkflowType string
	Input        any
	Config       WorkflowConfig
	TaskQueue    string
	Description  string
	StartAt      time.Time
	EndAt        time.Time
}

// CreateWorkflowSchedule creates a Workflow schedule with the specified configuration.
func (rs *WorkflowScheduler) CreateWorkflowSchedule(
	ctx context.Context,
	req WorkflowScheduleRequest,
) (client.ScheduleHandle, error) {
	if req.ID == "" {
		return nil, errors.New("schedule ID is required")
	}

	if req.WorkflowType == "" {
		return nil, errors.New("workflow type is required")
	}

	if len(req.Config.ExecutionTimes) == 0 {
		return nil, errors.New("workflow execution times are required")
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
		WithStartAt(req.StartAt).
		WithEndAt(req.EndAt).
		WithDescription(req.Description).
		WithCalendarSpec(executionTimes)

	// Set workflow action with task queue if provided
	workflowOptions := &client.StartWorkflowOptions{
		ID: req.ID,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		TaskQueue: req.TaskQueue,
	}

	builder = builder.WithWorkflowAction(req.WorkflowType, req.Input, workflowOptions)

	options, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build schedule configuration: %w", err)
	}

	return rs.manager.Create(ctx, options)
}

// CancelWorkflowSchedule cancels an active Workflow schedule.
func (rs *WorkflowScheduler) CancelWorkflowSchedule(ctx context.Context, scheduleID string) error {
	return rs.manager.Delete(ctx, scheduleID)
}

// PauseWorkflowSchedule pauses a Workflow schedule.
func (rs *WorkflowScheduler) PauseWorkflowSchedule(
	ctx context.Context,
	scheduleID, reason string,
) error {
	return rs.manager.Pause(ctx, scheduleID, reason)
}

// ResumeWorkflowSchedule resumes a paused Workflow schedule.
func (rs *WorkflowScheduler) ResumeWorkflowSchedule(
	ctx context.Context,
	scheduleID, reason string,
) error {
	return rs.manager.Resume(ctx, scheduleID, reason)
}

// GetWorkflowScheduleInfo retrieves information about a Workflow schedule.
func (rs *WorkflowScheduler) GetWorkflowScheduleInfo(
	ctx context.Context,
	scheduleID string,
) (*client.ScheduleDescription, error) {
	return rs.manager.Describe(ctx, scheduleID)
}
