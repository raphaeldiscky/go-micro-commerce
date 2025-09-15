package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"go.temporal.io/sdk/activity"

	pkgtemporal "github.com/raphaeldiscky/go-micro-commerce/pkg/temporal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
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

	// Initialize reminder scheduler first
	scheduleManager := pkgtemporal.NewTemporalScheduleManager(temporalClient.Client)
	reminderScheduler := pkgtemporal.NewReminderScheduler(scheduleManager)

	// Initialize payment reminder service
	paymentReminderService := service.NewPaymentReminderService(reminderScheduler)

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
		paymentReminderService,
	)

	// Register workflows
	temporalClient.Worker.RegisterWorkflow(temporal.OrderSagaWorkflow)
	temporalClient.Worker.RegisterWorkflow(temporal.PaymentReminderWorkflow)

	// Create payment reminder activities
	reminderActivities := temporal.NewPaymentReminderActivities(providers.DataStore)

	// Register order saga activities
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.ReserveProducts,
		activity.RegisterOptions{Name: string(constant.ReserveProductsStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.GetShippingCost,
		activity.RegisterOptions{Name: string(constant.GetShippingCostStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.SetFinalOrderPrices,
		activity.RegisterOptions{Name: string(constant.SetFinalPricesStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.CreatePayment,
		activity.RegisterOptions{Name: string(constant.CreatePaymentStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.SendPaymentRequiredNotification,
		activity.RegisterOptions{Name: string(constant.SendPaymentRequiredNotificationStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.WaitForPaymentConfirmation,
		activity.RegisterOptions{Name: string(constant.WaitForPaymentConfirmationStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.ProcessFulfillment,
		activity.RegisterOptions{Name: string(constant.ProcessFulfillmentStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.ConfirmProductsDeduction,
		activity.RegisterOptions{Name: string(constant.ConfirmProductsDeductionStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.SendOrderConfirmedNotification,
		activity.RegisterOptions{Name: string(constant.SendOrderConfirmedNotificationStep)},
	)

	// Register payment reminder activities
	temporalClient.Worker.RegisterActivityWithOptions(
		reminderActivities.SendPaymentReminderActivity,
		activity.RegisterOptions{Name: string(constant.SendPaymentReminderActivity)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		reminderActivities.CheckPaymentStatusActivity,
		activity.RegisterOptions{Name: string(constant.CheckPaymentStatusActivity)},
	)

	// Compensations
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.ReleaseProducts,
		activity.RegisterOptions{Name: string(constant.ReleaseProductsStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.RefundPayment,
		activity.RegisterOptions{Name: string(constant.RefundPaymentStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.RestoreProducts,
		activity.RegisterOptions{Name: string(constant.RestoreProductsStep)},
	)
	temporalClient.Worker.RegisterActivityWithOptions(
		activities.CancelShipping,
		activity.RegisterOptions{Name: string(constant.CancelShippingStep)},
	)

	// Set reminder scheduler in providers
	providers.ReminderScheduler = reminderScheduler
	providers.PaymentReminderService = paymentReminderService

	appLogger.Infof("Temporal client initialized with task queue: %s", cfg.Temporal.TaskQueue)

	return temporalClient
}
