package constant

import "time"

const (
	// SagaDefaultExecutionTimeout is the default execution timeout for sagas.
	SagaDefaultExecutionTimeout = 20 * time.Minute
	// SagaMaxConcurrent is the maximum number of concurrent sagas.
	SagaMaxConcurrent = 1000
	// SagaDefaultMaxRetries is the default maximum number of retries for sagas.
	SagaDefaultMaxRetries = 3
	// SagaDefaultRetryDelay is the default delay between retries for sagas.
	SagaDefaultRetryDelay = 2 * time.Second
	// SagaMaxRetryDelay is the maximum delay between retries for sagas.
	SagaMaxRetryDelay = 1 * time.Minute
	// SagaRecoveryInterval is the interval at which the recovery job runs.
	SagaRecoveryInterval = 5 * time.Minute
	// SagaRecoveryBatchSize is the number of sagas to process in a single recovery job run.
	SagaRecoveryBatchSize = 100
	// SagaMaxRecoveryAge is the maximum age for which a saga can be recovered.
	SagaMaxRecoveryAge = 24 * time.Hour
	// SagaStateRetention is the time period for which saga states are retained.
	SagaStateRetention = 30 * 24 * time.Hour
	// SagaPurgeInterval is the interval at which the saga state purge job runs.
	SagaPurgeInterval = 24 * time.Hour
)

// OrderSagaWorkflowName defines the workflow name.
const OrderSagaWorkflowName = "OrderSagaWorkflow"

// SagaStatus represents the status of a saga execution.
type SagaStatus string

const (
	// SagaStatusPending indicates that the saga is pending.
	SagaStatusPending SagaStatus = "pending"
	// SagaStatusExecuting indicates that the saga is currently executing.
	SagaStatusExecuting SagaStatus = "executing"
	// SagaStatusCompensating indicates that the saga is compensating due to a failure.
	SagaStatusCompensating SagaStatus = "compensating"
	// SagaStatusCompleted indicates that the saga has completed successfully.
	SagaStatusCompleted SagaStatus = "completed"
	// SagaStatusFailed indicates that the saga has failed.
	SagaStatusFailed SagaStatus = "failed"
	// SagaStatusCompensated indicates that the saga has been compensated after a failure.
	SagaStatusCompensated SagaStatus = "compensated"
)

// PaymentStatus indicates the status of waiting for payment.
type PaymentStatus string

const (
	// PaymentStatusTimeout indicates that the payment is pending.
	PaymentStatusTimeout PaymentStatus = "timeout"
	// PaymentStatusCompleted indicates that the payment has completed.
	PaymentStatusCompleted PaymentStatus = "completed"
	// PaymentStatusFailed indicates that the payment has failed.
	PaymentStatusFailed PaymentStatus = "failed"
)

// WorkflowStep represents a step in the workflow.
type WorkflowStep string

// ====== Step 1 ======.
const (
	// ReserveProductsStep reserves the products in the order.
	ReserveProductsStep WorkflowStep = "ReserveProducts"
	// ReserveProductsStepMaxRetries is the maximum number of retries for the ReserveProductsStep.
	ReserveProductsStepMaxRetries int = 3
	// ReserveProductsStepRetryDelay is the delay between retries for the ReserveProductsStep.
	ReserveProductsStepRetryDelay = 500 * time.Millisecond
	// ReserveProductsStepTimeout is the timeout for the ReserveProductsStep.
	ReserveProductsStepTimeout = 30 * time.Second
	// ReleaseProductsStep releases the reserved products.
	ReleaseProductsStep WorkflowStep = "ReleaseProducts"
)

// ====== Step 2 ======.
const (
	// GetShippingCostStep gets the shipping cost for the order.
	GetShippingCostStep WorkflowStep = "GetShippingCost"
	// GetShippingCostStepMaxRetries is the maximum number of retries for the GetShippingCostStep.
	GetShippingCostStepMaxRetries int = 2
	// GetShippingCostStepRetryDelay is the delay between retries for the GetShippingCostStep.
	GetShippingCostStepRetryDelay = 3 * time.Second
	// GetShippingCostStepTimeout is the timeout for the GetShippingCostStep.
	GetShippingCostStepTimeout = 30 * time.Second
)

// ====== Step 3 ======.
const (
	// SetFinalPricesStep set the final prices for the order (+shipping cost).
	SetFinalPricesStep WorkflowStep = "SetFinalOrderPrices"
	// SetFinalPricesStepMaxRetries is the maximum number of retries for the SetFinalPricesStep.
	SetFinalPricesStepMaxRetries int = 3
	// SetFinalPricesStepRetryDelay is the delay between retries for the SetFinalPricesStep.
	SetFinalPricesStepRetryDelay = 1 * time.Second
	// SetFinalPricesStepTimeout is the timeout for the SetFinalPricesStep.
	SetFinalPricesStepTimeout = 10 * time.Second
)

// ====== Step 4 ======.
const (
	// CreatePaymentStep processes the payment for the order.
	CreatePaymentStep WorkflowStep = "CreatePayment"
	// RefundPaymentStep refunds the payment.
	RefundPaymentStep WorkflowStep = "RefundPayment"
	// CreatePaymentStepMaxRetries is the maximum number of retries for the CreatePaymentStep.
	CreatePaymentStepMaxRetries int = 3
	// CreatePaymentStepRetryDelay is the delay between retries for the CreatePaymentStep.
	CreatePaymentStepRetryDelay = 5 * time.Second
	// CreatePaymentStepTimeout is the timeout for the CreatePaymentStep.
	CreatePaymentStepTimeout = 30 * time.Second
)

// ====== Step 5 ======.
const (
	// SendPaymentRequiredNotificationStep sends a payment required notification.
	SendPaymentRequiredNotificationStep WorkflowStep = "SendPaymentRequiredNotification"
	// SendPaymentRequiredNotificationStepMaxRetries is the maximum number of retries for the SendPaymentRequiredNotificationStep.
	SendPaymentRequiredNotificationStepMaxRetries int = 3
	// SendPaymentRequiredNotificationStepRetryDelay is the delay between retries for the SendPaymentRequiredNotificationStep.
	SendPaymentRequiredNotificationStepRetryDelay = 5 * time.Second
	// SendPaymentRequiredNotificationStepTimeout is the timeout for the SendPaymentRequiredNotificationStep.
	SendPaymentRequiredNotificationStepTimeout = 60 * time.Second
)

// ====== Step 6 ======.
const (
	// WaitForPaymentConfirmationStep waits for the payment confirmation.
	WaitForPaymentConfirmationStep WorkflowStep = "WaitForPaymentConfirmation"
	// WaitForPaymentConfirmationStepMaxRetries is the maximum number of retries for the WaitForPaymentConfirmationStep.
	WaitForPaymentConfirmationStepMaxRetries int = 3
	// WaitForPaymentConfirmationStepRetryDelay is the delay between retries for the WaitForPaymentConfirmationStep.
	WaitForPaymentConfirmationStepRetryDelay = 5 * time.Second
	// WaitForPaymentConfirmationStepTimeout is the timeout for the WaitForPaymentConfirmationStep.
	WaitForPaymentConfirmationStepTimeout = 5 * time.Minute
	// ExtendedPaymentTimeout is the extended timeout while reminders are active.
	ExtendedPaymentTimeout = 30 * time.Minute
)

// ====== Step 7 ======.
const (
	// ProcessFulfillmentStep creates a shipping order.
	ProcessFulfillmentStep WorkflowStep = "ProcessFulfillment"
	// ProcessFulfillmentStepMaxRetries is the maximum number of retries for the ProcessFulfillmentStep.
	ProcessFulfillmentStepMaxRetries int = 2
	// ProcessFulfillmentStepRetryDelay is the delay between retries for the ProcessFulfillmentStep.
	ProcessFulfillmentStepRetryDelay = 3 * time.Second
	// ProcessFulfillmentStepTimeout is the timeout for the ProcessFulfillmentStep.
	ProcessFulfillmentStepTimeout = 30 * time.Second
	// CancelShippingStep cancels the shipping order.
	CancelShippingStep WorkflowStep = "CancelShipping"
)

// ====== Step 8 ======.
const (
	// ConfirmProductsDeductionStep deducts the products from inventory.
	ConfirmProductsDeductionStep WorkflowStep = "ConfirmProductsDeduction"
	// ConfirmProductsDeductionStepMaxRetries is the maximum number of retries for the ConfirmProductsDeductionStep.
	ConfirmProductsDeductionStepMaxRetries int = 3
	// ConfirmProductsDeductionStepRetryDelay is the delay between retries for the ConfirmProductsDeductionStep.
	ConfirmProductsDeductionStepRetryDelay = 2 * time.Second
	// ConfirmProductsDeductionStepTimeout is the timeout for the ConfirmProductsDeductionStep.
	ConfirmProductsDeductionStepTimeout = 20 * time.Second
	// RestoreProductsStep compensation for restoring the products.
	RestoreProductsStep WorkflowStep = "RestoreProducts"
)

// ====== Step 9 ======.
const (
	// SendOrderConfirmedNotificationStep sends an order confirmation.
	SendOrderConfirmedNotificationStep WorkflowStep = "SendOrderConfirmedNotification"
	// SendOrderConfirmedNotificationStepMaxRetries is the maximum number of retries for the SendOrderConfirmedNotificationStep.
	SendOrderConfirmedNotificationStepMaxRetries int = 3
	// SendOrderConfirmedNotificationStepRetryDelay is the delay between retries for the SendOrderConfirmedNotificationStep.
	SendOrderConfirmedNotificationStepRetryDelay = 1 * time.Second
	// SendOrderConfirmedNotificationStepTimeout is the timeout for the SendOrderConfirmedNotificationStep.
	SendOrderConfirmedNotificationStepTimeout = 10 * time.Second
)
