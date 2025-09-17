// Package job provides background job for the order service.
package job

import (
	"context"
	"sync"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
)

// Job defines the methods for a job.
type Job interface {
	Start(ctx context.Context)
	Stop()
	Name() string
	Interval() time.Duration
	IsEnabled() bool
}

// Scheduler manages and coordinates multiple jobs.
type Scheduler struct {
	jobs   []Job
	logger logger.Logger
	config *config.JobConfig
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewScheduler creates a new job scheduler.
func NewScheduler(appLogger logger.Logger, config *config.JobConfig) *Scheduler {
	return &Scheduler{
		jobs:   make([]Job, 0),
		logger: appLogger,
		config: config,
	}
}

// RegisterJob adds a job to the scheduler.
func (s *Scheduler) RegisterJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs = append(s.jobs, job)
	s.logger.Infof("Registered job: %s (interval: %v, enabled: %v)",
		job.Name(), job.Interval(), job.IsEnabled())
}

// Start begins executing all registered job.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.logger.Infof("Starting job scheduler with %d registered job", len(s.jobs))

	// Start each enabled job in its own goroutine
	for _, job := range s.jobs {
		if job.IsEnabled() {
			s.wg.Add(1)

			go s.runJob(s.ctx, job)
		} else {
			s.logger.Infof("Job %s is disabled, skipping", job.Name())
		}
	}

	// Wait for context cancellation
	<-s.ctx.Done()
	s.logger.Info("Job scheduler received shutdown signal")

	// Stop all job
	s.stopAllJob()

	// Wait for all job to finish with timeout
	done := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All job stopped gracefully")
	case <-time.After(s.config.Recovery.Timeout):
		s.logger.Warn("Job shutdown timeout reached, some job may not have stopped gracefully")
	}

	return nil
}

// runJob executes a single job with its configured interval.
func (s *Scheduler) runJob(ctx context.Context, job Job) {
	defer s.wg.Done()

	s.logger.Infof("Starting job: %s with interval %v", job.Name(), job.Interval())

	ticker := time.NewTicker(job.Interval())
	defer ticker.Stop()

	// Execute immediately on start
	if job.IsEnabled() {
		s.executeJob(ctx, job)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("Job %s received shutdown signal", job.Name())
			job.Stop()

			return
		case <-ticker.C:
			if job.IsEnabled() {
				s.executeJob(ctx, job)
			}
		}
	}
}

// executeJob runs a single job execution with error handling.
func (s *Scheduler) executeJob(ctx context.Context, job Job) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("Job %s panicked: %v", job.Name(), r)
		}
	}()

	s.logger.Debugf("Executing job: %s", job.Name())

	start := time.Now()

	job.Start(ctx)

	duration := time.Since(start)

	s.logger.Debugf("Job %s completed in %v", job.Name(), duration)
}

// stopAllJob gracefully stops all job.
func (s *Scheduler) stopAllJob() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Info("Stopping all job...")

	for _, job := range s.jobs {
		job.Stop()
	}
}

// Shutdown stops the scheduler and all job.
func (s *Scheduler) Shutdown(_ context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

// GetJobtatus returns the status of all registered job.
func (s *Scheduler) GetJobtatus() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]any)
	for _, job := range s.jobs {
		status[job.Name()] = map[string]any{
			"enabled":  job.IsEnabled(),
			"interval": job.Interval().String(),
		}
	}

	return status
}

// GetJobCount returns the number of registered job.
func (s *Scheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.jobs)
}
