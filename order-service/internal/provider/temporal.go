package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/temporal"
)

// SetupTemporal initializes the Temporal client.
func SetupTemporal(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *client.TemporalClient {
	// Setup product client
	productClient, err := client.NewProductClient(cfg)
	if err != nil {
		appLogger.Warnf(
			"failed to create product client: %v. Temporal workflows will start without product client functionality.",
			err,
		)

		productClient = nil
	}

	// Create Temporal client
	temporalClient, err := client.NewTemporalClient(
		cfg.Temporal,
		appLogger,
	)
	if err != nil {
		appLogger.Warnf("failed to create Temporal client: %v", err)

		return nil
	}

	// Create temporal activities and register them
	// Note: Temporal activities use simplified approach for event handling
	activities := temporal.NewTemporalActivities(
		providers.DataStore,
		productClient,
		nil, // paymentRequestProducer - not needed for Temporal direct approach
		nil, // orderLifecycleProducer - not needed for Temporal direct approach
		nil, // fulfillmentRequestProducer - not needed for Temporal direct approach
		providers.FulfillmentClient,
		providers.PaymentClient,
	)

	// Register workflow and activities
	temporalClient.Worker.RegisterWorkflow(temporal.OrderSagaWorkflow)
	temporalClient.Worker.RegisterActivity(activities.ReserveProductsAndCalculate)
	temporalClient.Worker.RegisterActivity(activities.ProcessFulfillment)
	temporalClient.Worker.RegisterActivity(activities.SetFinalOrderPrices)
	temporalClient.Worker.RegisterActivity(activities.ProcessPayment)
	temporalClient.Worker.RegisterActivity(activities.ConfirmProductsDeduction)
	temporalClient.Worker.RegisterActivity(activities.SendOrderConfirmation)
	temporalClient.Worker.RegisterActivity(activities.ReleaseProducts)
	temporalClient.Worker.RegisterActivity(activities.RefundPayment)
	temporalClient.Worker.RegisterActivity(activities.RestoreProducts)
	temporalClient.Worker.RegisterActivity(activities.CancelShipping)

	appLogger.Infof("Temporal client initialized with task queue: %s", cfg.Temporal.TaskQueue)

	return temporalClient
}
