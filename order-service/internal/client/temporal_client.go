package client

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
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
	appLogger logger.Logger,
) (*TemporalClient, error) {
	clientOptions := client.Options{
		HostPort:    cfg.Address,
		Namespace:   cfg.Namespace,
		Credentials: client.NewAPIKeyStaticCredentials(cfg.APIKey),
	}

	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	appLogger.Info("Successfully created Temporal client")

	w := worker.New(temporalClient, cfg.TaskQueue, worker.Options{})

	return &TemporalClient{
		Client: temporalClient,
		Worker: w,
		config: cfg,
	}, nil
}

// CreateOrderWorkflowOptions returns the options to create a Temporal workflow.
func (tc *TemporalClient) CreateOrderWorkflowOptions(
	orderID uuid.UUID,
) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		TaskQueue: tc.config.TaskQueue,
		ID:        sagautils.CreateOrderSagaID(orderID),
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: tc.config.BackoffCoefficient,
			MaximumInterval:    tc.config.MaxInterval,
			MaximumAttempts:    tc.config.MaxAttempts,
		},
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
