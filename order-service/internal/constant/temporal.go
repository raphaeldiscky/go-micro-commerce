package constant

import "time"

const (
	// TemporalRetryInterval is the retry interval for Temporal tasks.
	TemporalRetryInterval = 1 * time.Second
	// TemporalBackoffCoefficient is the backoff coefficient for Temporal tasks.
	TemporalBackoffCoefficient = 2.0
	// TemporalMaxAttempts is the maximum number of attempts for Temporal tasks.
	TemporalMaxAttempts = 3
	// TemporalMaxInterval is the maximum interval for Temporal tasks.
	TemporalMaxInterval = 1 * time.Minute
	// TemporalWorkflowTimeout is the start-to-close timeout for Temporal tasks.
	TemporalWorkflowTimeout = 20 * time.Minute
	// TemporalCompensationWorkflowTimeout is the start-to-close timeout for compensation Temporal tasks.
	TemporalCompensationWorkflowTimeout = 15 * time.Minute
)

// Temporal Activity Timeouts and Retry Policies (matching saga step constants).
const (
	// ReserveProductsActivityTimeout is the timeout for ReserveProducts activity.
	ReserveProductsActivityTimeout = ReserveProductsStepTimeout
	// ReserveProductsActivityMaxRetries is the maximum number of retries for ReserveProducts activity.
	ReserveProductsActivityMaxRetries = ReserveProductsStepMaxRetries
	// ReserveProductsActivityRetryInterval is the retry interval for ReserveProducts activity.
	ReserveProductsActivityRetryInterval = ReserveProductsStepRetryDelay

	// GetShippingCostActivityTimeout is the timeout for GetShippingCost activity.
	GetShippingCostActivityTimeout = GetShippingCostStepTimeout
	// GetShippingCostActivityMaxRetries is the maximum number of retries for GetShippingCost activity.
	GetShippingCostActivityMaxRetries = GetShippingCostStepMaxRetries
	// GetShippingCostActivityRetryInterval is the retry interval for GetShippingCost activity.
	GetShippingCostActivityRetryInterval = GetShippingCostStepRetryDelay

	// SetFinalPricesActivityTimeout is the timeout for SetFinalPrices activity.
	SetFinalPricesActivityTimeout = SetFinalPricesStepTimeout
	// SetFinalPricesActivityMaxRetries is the maximum number of retries for SetFinalPrices activity.
	SetFinalPricesActivityMaxRetries = SetFinalPricesStepMaxRetries
	// SetFinalPricesActivityRetryInterval is the retry interval for SetFinalPrices activity.
	SetFinalPricesActivityRetryInterval = SetFinalPricesStepRetryDelay

	// CreatePaymentActivityTimeout is the timeout for CreatePayment activity.
	CreatePaymentActivityTimeout = CreatePaymentStepTimeout
	// CreatePaymentActivityMaxRetries is the maximum number of retries for CreatePayment activity.
	CreatePaymentActivityMaxRetries = CreatePaymentStepMaxRetries
	// CreatePaymentActivityRetryInterval is the retry interval for CreatePayment activity.
	CreatePaymentActivityRetryInterval = CreatePaymentStepRetryDelay

	// SendPaymentRequiredNotificationActivityTimeout is the timeout for SendPaymentRequiredNotification activity.
	SendPaymentRequiredNotificationActivityTimeout = SendPaymentRequiredNotificationStepTimeout
	// SendPaymentRequiredNotificationActivityMaxRetries is the maximum number of retries for SendPaymentRequiredNotification activity.
	SendPaymentRequiredNotificationActivityMaxRetries = SendPaymentRequiredNotificationStepMaxRetries
	// SendPaymentRequiredNotificationActivityRetryInterval is the retry interval for SendPaymentRequiredNotification activity.
	SendPaymentRequiredNotificationActivityRetryInterval = SendPaymentRequiredNotificationStepRetryDelay

	// WaitForPaymentConfirmationActivityTimeout is the timeout for WaitForPaymentConfirmation activity.
	WaitForPaymentConfirmationActivityTimeout = WaitForPaymentConfirmationStepTimeout
	// WaitForPaymentConfirmationActivityMaxRetries is the maximum number of retries for WaitForPaymentConfirmation activity.
	WaitForPaymentConfirmationActivityMaxRetries = WaitForPaymentConfirmationStepMaxRetries
	// WaitForPaymentConfirmationActivityRetryInterval is the retry interval for WaitForPaymentConfirmation activity.
	WaitForPaymentConfirmationActivityRetryInterval = WaitForPaymentConfirmationStepRetryDelay

	// ProcessFulfillmentActivityTimeout is the timeout for ProcessFulfillment activity.
	ProcessFulfillmentActivityTimeout = ProcessFulfillmentStepTimeout
	// ProcessFulfillmentActivityMaxRetries is the maximum number of retries for ProcessFulfillment activity.
	ProcessFulfillmentActivityMaxRetries = ProcessFulfillmentStepMaxRetries
	// ProcessFulfillmentActivityRetryInterval is the retry interval for ProcessFulfillment activity.
	ProcessFulfillmentActivityRetryInterval = ProcessFulfillmentStepRetryDelay

	// ConfirmProductsDeductionActivityTimeout is the timeout for ConfirmProductsDeduction activity.
	ConfirmProductsDeductionActivityTimeout = ConfirmProductsDeductionStepTimeout
	// ConfirmProductsDeductionActivityMaxRetries is the maximum number of retries for ConfirmProductsDeduction activity.
	ConfirmProductsDeductionActivityMaxRetries = ConfirmProductsDeductionStepMaxRetries
	// ConfirmProductsDeductionActivityRetryInterval is the retry interval for ConfirmProductsDeduction activity.
	ConfirmProductsDeductionActivityRetryInterval = ConfirmProductsDeductionStepRetryDelay

	// SendOrderConfirmedNotificationActivityTimeout is the timeout for SendOrderConfirmedNotification activity.
	SendOrderConfirmedNotificationActivityTimeout = SendOrderConfirmedNotificationStepTimeout
	// SendOrderConfirmedNotificationActivityMaxRetries is the maximum number of retries for SendOrderConfirmedNotification activity.
	SendOrderConfirmedNotificationActivityMaxRetries = SendOrderConfirmedNotificationStepMaxRetries
	// SendOrderConfirmedNotificationActivityRetryInterval is the retry interval for SendOrderConfirmedNotification activity.
	SendOrderConfirmedNotificationActivityRetryInterval = SendOrderConfirmedNotificationStepRetryDelay
)
