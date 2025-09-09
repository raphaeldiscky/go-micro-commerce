package constant

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

// WorkflowStep represents a step in the workflow.
type WorkflowStep string

const (
	// ReserveProductsStep reserves the products in the order.
	ReserveProductsStep WorkflowStep = "ReserveProducts"
	// GetShippingCostStep gets the shipping cost for the order.
	GetShippingCostStep WorkflowStep = "GetShippingCost"
	// SetFinalPricesStep set the final prices for the order (+shipping cost).
	SetFinalPricesStep WorkflowStep = "SetFinalOrderPrices"
	// CreatePaymentStep processes the payment for the order.
	CreatePaymentStep WorkflowStep = "CreatePayment"
	// SendPaymentRequiredNotificationStep sends a payment required notification.
	SendPaymentRequiredNotificationStep WorkflowStep = "SendPaymentRequiredNotification"
	// WaitForPaymentConfirmationStep waits for the payment confirmation.
	WaitForPaymentConfirmationStep WorkflowStep = "WaitForPaymentConfirmation"
	// ProcessFulfillmentStep creates a shipping order.
	ProcessFulfillmentStep WorkflowStep = "ProcessFulfillment"
	// ConfirmProductsDeductionStep deducts the products from inventory.
	ConfirmProductsDeductionStep WorkflowStep = "ConfirmProductsDeduction"
	// SendOrderConfirmedNotificationStep sends an order confirmation.
	SendOrderConfirmedNotificationStep WorkflowStep = "SendOrderConfirmedNotification"
)

const (
	// ReleaseProductsStep releases the reserved products.
	ReleaseProductsStep WorkflowStep = "ReleaseProducts"
	// RefundPaymentStep refunds the payment.
	RefundPaymentStep WorkflowStep = "RefundPayment"
	// RestoreProductsStep restores the reserved products.
	RestoreProductsStep WorkflowStep = "RestoreProducts"
	// CancelShippingStep cancels the shipping order.
	CancelShippingStep WorkflowStep = "CancelShipping"
)
