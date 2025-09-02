package client

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/sagautils"
)

// TemporalClient holds Temporal-related components.
type TemporalClient struct {
	config *config.TemporalConfig
	Client client.Client
	Worker worker.Worker
}

// NewTemporalClient creates and configures a Temporal client.
func NewTemporalClient(
	cfg *config.TemporalConfig,
	_ logger.Logger,
) (*TemporalClient, error) {
	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	// Create worker (activities will be registered separately)
	w := worker.New(temporalClient, cfg.TaskQueue, worker.Options{})

	return &TemporalClient{
		Client: temporalClient,
		Worker: w,
		config: cfg,
	}, nil
}

// CreateWorkflowOptions returns the options to create a Temporal workflow.
func (tc *TemporalClient) CreateWorkflowOptions(orderID uuid.UUID) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		TaskQueue: tc.config.TaskQueue,
		ID:        sagautils.CreateOrderSagaID(orderID),
	}
}

// Close shuts down the Temporal client.
func (tc *TemporalClient) Close() {
	if tc.Worker != nil {
		tc.Worker.Stop()
	}

	if tc.Client != nil {
		tc.Client.Close()
	}
}
