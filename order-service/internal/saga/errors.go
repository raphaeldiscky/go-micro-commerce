package saga

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// ErrorType defines different types of saga errors.
type ErrorType string

const (
	// ErrorTypeRetriable indicates an error that can be retried.
	ErrorTypeRetriable ErrorType = "retriable"
	// ErrorTypeNonRetriable indicates an error that should not be retried.
	ErrorTypeNonRetriable ErrorType = "non_retriable"
	// ErrorTypeTimeout indicates a timeout error.
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeCancellation indicates a cancellation error.
	ErrorTypeCancellation ErrorType = "cancellation"
	// ErrorTypeBusinessRule indicates a business rule violation.
	ErrorTypeBusinessRule ErrorType = "business_rule"
)

// Error represents a structured error in saga execution.
type Error struct {
	Type    ErrorType
	Message string
	Cause   error
	Step    constant.WorkflowStep
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Step, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s: %s", e.Step, e.Message)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// IsRetriable returns true if the error can be retried.
func (e *Error) IsRetriable() bool {
	return e.Type == ErrorTypeRetriable
}

// NewRetriableError creates a new retriable error.
func NewRetriableError(step constant.WorkflowStep, message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeRetriable,
		Message: message,
		Cause:   cause,
		Step:    step,
	}
}

// NewNonRetriableError creates a new non-retriable error.
func NewNonRetriableError(step constant.WorkflowStep, message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeNonRetriable,
		Message: message,
		Cause:   cause,
		Step:    step,
	}
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(step constant.WorkflowStep, message string, cause error) *Error {
	return &Error{
		Type:    ErrorTypeTimeout,
		Message: message,
		Cause:   cause,
		Step:    step,
	}
}

// CategorizeError categorizes an error based on its type and context.
func CategorizeError(step constant.WorkflowStep, err error) *Error {
	if err == nil {
		return nil
	}

	// Check if it's already a Error
	var sagaErr *Error
	if errors.As(err, &sagaErr) {
		return sagaErr
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) {
		return &Error{
			Type:    ErrorTypeCancellation,
			Message: "operation was canceled",
			Cause:   err,
			Step:    step,
		}
	}

	// Check for timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{
			Type:    ErrorTypeTimeout,
			Message: "operation timed out",
			Cause:   err,
			Step:    step,
		}
	}

	// Default to retriable for unknown errors
	return &Error{
		Type:    ErrorTypeRetriable,
		Message: "unknown error occurred",
		Cause:   err,
		Step:    step,
	}
}

// isTemporaryError checks if an error is temporary and can be retried.
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common temporary error patterns
	errMsg := err.Error()

	// Database connection errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "temporary failure") ||
		strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "rate limit") {
		return true
	}

	// HTTP status code errors (5xx are temporary)
	if strings.Contains(errMsg, "500") ||
		strings.Contains(errMsg, "502") ||
		strings.Contains(errMsg, "503") ||
		strings.Contains(errMsg, "504") {
		return true
	}

	return false
}
