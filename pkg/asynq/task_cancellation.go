package asynq

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

// TaskCancellationService handles cancellation of scheduled tasks for cleanup.
type TaskCancellationService interface {
	CancelTasksByPattern(ctx context.Context, queue string, taskIDPattern string) error
	CancelTask(ctx context.Context, queue string, taskID string) error
}

type taskCancellationService struct {
	inspector *asynq.Inspector
}

// NewTaskCancellationService creates a new task cancellation service.
func NewTaskCancellationService(inspector *asynq.Inspector) TaskCancellationService {
	return &taskCancellationService{
		inspector: inspector,
	}
}

// CancelTask cancels a specific task by its ID.
func (s *taskCancellationService) CancelTask(
	_ context.Context,
	queue string,
	taskID string,
) error {
	if err := s.inspector.DeleteTask(queue, taskID); err != nil {
		return fmt.Errorf("failed to cancel task %s in queue %s: %w", taskID, queue, err)
	}

	return nil
}

// CancelTasksByPattern cancels tasks matching a pattern (for bulk operations).
func (s *taskCancellationService) CancelTasksByPattern(
	_ context.Context,
	queue string,
	taskIDPattern string,
) error {
	// List all tasks in the queue
	tasks, err := s.inspector.ListScheduledTasks(queue)
	if err != nil {
		return fmt.Errorf("failed to list tasks in queue %s: %w", queue, err)
	}

	// Cancel tasks matching the pattern
	for _, task := range tasks {
		if matchesPattern(task.ID, taskIDPattern) {
			if errTask := s.inspector.DeleteTask(queue, task.ID); errTask != nil {
				// Log error but continue with other tasks
				continue
			}
		}
	}

	return nil
}

// matchesPattern checks if a task ID matches the given pattern.
func matchesPattern(taskID, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex if needed
	// For now, check if pattern is a prefix of taskID
	return len(taskID) >= len(pattern) && taskID[:len(pattern)] == pattern
}
