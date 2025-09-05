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
	// ReserveProductsAndCalculateStep reserves the products in the order.
	ReserveProductsAndCalculateStep WorkflowStep = "ReserveProductsAndCalculate"
	// ProcessPaymentStep processes the payment for the order.
	ProcessPaymentStep WorkflowStep = "ProcessPayment"
	// ConfirmProductsDeductionStep deducts the products from inventory.
	ConfirmProductsDeductionStep WorkflowStep = "ConfirmProductsDeduction"
	// ProcessFulfillmentStep creates a shipping order.
	ProcessFulfillmentStep WorkflowStep = "ProcessFulfillment"
	// SendOrderConfirmationStep sends an order confirmation.
	SendOrderConfirmationStep WorkflowStep = "SendOrderConfirmation"
)
