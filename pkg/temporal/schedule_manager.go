package temporal

import (
	"context"
	"errors"

	"go.temporal.io/sdk/client"
)

// ScheduleManager provides interface for managing Temporal schedules.
type ScheduleManager interface {
	// Create creates a new schedule.
	Create(ctx context.Context, options ScheduleOptions) (client.ScheduleHandle, error)
	// Pause pauses a schedule.
	Pause(ctx context.Context, scheduleID string, note string) error
	// Resume resumes a paused schedule.
	Resume(ctx context.Context, scheduleID string, note string) error
	// Delete deletes a schedule.
	Delete(ctx context.Context, scheduleID string) error
	// Get retrieves a schedule handle.
	Get(ctx context.Context, scheduleID string) (client.ScheduleHandle, error)
	// Describe describes a schedule.
	Describe(ctx context.Context, scheduleID string) (*client.ScheduleDescription, error)
	// List lists all schedules.
	List(ctx context.Context, query string) (client.ScheduleListIterator, error)
	// Trigger triggers a schedule immediately.
	Trigger(ctx context.Context, scheduleID string) error
}

// scheduleManager implements ScheduleManager using Temporal client.
type scheduleManager struct {
	client client.Client
}

// NewTemporalScheduleManager creates a new scheduleManager.
func NewTemporalScheduleManager(temporalClient client.Client) ScheduleManager {
	return &scheduleManager{
		client: temporalClient,
	}
}

// Create creates a new schedule.
func (m *scheduleManager) Create(
	ctx context.Context,
	options ScheduleOptions,
) (client.ScheduleHandle, error) {
	if options.ID == "" {
		return nil, errors.New("schedule ID is required")
	}

	if options.Action == nil {
		return nil, errors.New("schedule action is required")
	}

	clientOptions := client.ScheduleOptions{
		ID:     options.ID,
		Spec:   options.Spec,
		Action: options.Action,
	}

	return m.client.ScheduleClient().Create(ctx, clientOptions)
}

// Pause pauses a schedule.
func (m *scheduleManager) Pause(ctx context.Context, scheduleID string, note string) error {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)

	return handle.Pause(ctx, client.SchedulePauseOptions{
		Note: note,
	})
}

// Resume resumes a paused schedule.
func (m *scheduleManager) Resume(
	ctx context.Context,
	scheduleID string,
	note string,
) error {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)

	return handle.Unpause(ctx, client.ScheduleUnpauseOptions{
		Note: note,
	})
}

// Delete deletes a schedule.
func (m *scheduleManager) Delete(ctx context.Context, scheduleID string) error {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)
	return handle.Delete(ctx)
}

// Get retrieves a schedule handle.
func (m *scheduleManager) Get(
	ctx context.Context,
	scheduleID string,
) (client.ScheduleHandle, error) {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)
	return handle, nil
}

// Describe describes a schedule.
func (m *scheduleManager) Describe(
	ctx context.Context,
	scheduleID string,
) (*client.ScheduleDescription, error) {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)
	return handle.Describe(ctx)
}

// List lists all schedules.
func (m *scheduleManager) List(
	ctx context.Context,
	query string,
) (client.ScheduleListIterator, error) {
	return m.client.ScheduleClient().List(ctx, client.ScheduleListOptions{
		Query: query,
	})
}

// Trigger triggers a schedule immediately.
func (m *scheduleManager) Trigger(ctx context.Context, scheduleID string) error {
	handle := m.client.ScheduleClient().GetHandle(ctx, scheduleID)
	return handle.Trigger(ctx, client.ScheduleTriggerOptions{})
}
