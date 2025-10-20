package stripe

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v83"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/dto"
)

// MapPaymentIntentToResponse converts a Stripe PaymentIntent to PaymentGatewayResponse.
func MapPaymentIntentToResponse(
	pi *stripe.PaymentIntent,
	transactionID uuid.UUID,
) (*dto.PaymentGatewayResponse, error) {
	// Convert amount from cents to decimal
	amount := decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(multiplyAmount))

	// Map Stripe status to our internal status
	status := mapStripeStatusToGatewayStatus(string(pi.Status))

	// Set client_secret for frontend to confirm payment with Stripe.js
	clientSecret := pi.ClientSecret

	response := &dto.PaymentGatewayResponse{
		TransactionID: transactionID,
		GatewayID:     pi.ID,
		Status:        status,
		Amount:        amount,
		Currency:      string(pi.Currency),
		ProcessedAt:   time.Unix(pi.Created, 0),
		ClientSecret:  &clientSecret, // For stripe.confirmCardPayment() on frontend
		GatewayResponse: map[string]any{
			"status":         string(pi.Status),
			"client_secret":  pi.ClientSecret,
			"payment_method": pi.PaymentMethod,
		},
	}

	// Note: Fees are available after the charge is completed
	// They can be retrieved separately using the Charge API if needed

	// Add failure reason if failed
	if pi.Status == stripe.PaymentIntentStatusCanceled && pi.CancellationReason != "" {
		response.FailureReason = string(pi.CancellationReason)
	}

	// Check if requires action (3D Secure, etc.)
	if pi.Status == stripe.PaymentIntentStatusRequiresAction {
		response.RequiresAction = true
		if pi.NextAction != nil {
			response.NextAction = &dto.PaymentAction{
				Type: constant.PaymentActionTypeRedirect,
				URL:  pi.NextAction.RedirectToURL.URL,
			}
		}
	}

	return response, nil
}

// MapRefundToResponse converts a Stripe Refund to RefundResponse.
func MapRefundToResponse(
	r *stripe.Refund,
	refundID, transactionID uuid.UUID,
) (*dto.RefundResponse, error) {
	// Convert amount from cents to decimal
	amount := decimal.NewFromInt(r.Amount).Div(decimal.NewFromInt(multiplyAmount))

	// Map Stripe refund status to our internal status
	status := mapStripeRefundStatus(string(r.Status))

	response := &dto.RefundResponse{
		RefundID:        refundID,
		TransactionID:   transactionID,
		GatewayRefundID: r.ID,
		Status:          status,
		Amount:          amount,
		Currency:        string(r.Currency),
		ProcessedAt:     time.Unix(r.Created, 0),
	}

	// Add fees if available
	if r.BalanceTransaction != nil {
		feeAmount := decimal.NewFromInt(r.BalanceTransaction.Fee).
			Div(decimal.NewFromInt(multiplyAmount))
		response.Fees = &feeAmount
	}

	return response, nil
}

// mapStripeStatusToGatewayStatus maps Stripe PaymentIntent status to internal status.
func mapStripeStatusToGatewayStatus(stripeStatus string) constant.PaymentGatewayStatus {
	switch stripe.PaymentIntentStatus(stripeStatus) {
	case stripe.PaymentIntentStatusSucceeded:
		return constant.PaymentGatewayStatusSucceeded
	case stripe.PaymentIntentStatusProcessing:
		return constant.PaymentGatewayStatusPending
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return constant.PaymentGatewayStatusPending
	case stripe.PaymentIntentStatusRequiresConfirmation:
		return constant.PaymentGatewayStatusPending
	case stripe.PaymentIntentStatusRequiresAction:
		return constant.PaymentGatewayStatusPending
	case stripe.PaymentIntentStatusCanceled:
		return constant.PaymentGatewayStatusCanceled
	case stripe.PaymentIntentStatusRequiresCapture:
		return constant.PaymentGatewayStatusPending
	default:
		return constant.PaymentGatewayStatusFailed
	}
}

// mapStripeRefundStatus maps Stripe refund status to internal status.
func mapStripeRefundStatus(stripeStatus string) constant.RefundStatus {
	switch stripe.RefundStatus(stripeStatus) {
	case stripe.RefundStatusSucceeded:
		return constant.RefundStatusSucceeded
	case stripe.RefundStatusPending:
		return constant.RefundStatusPending
	case stripe.RefundStatusFailed:
		return constant.RefundStatusFailed
	case stripe.RefundStatusCanceled:
		return constant.RefundStatusCanceled
	case stripe.RefundStatusRequiresAction:
		return constant.RefundStatusPending
	default:
		return constant.RefundStatusFailed
	}
}

// parseTransactionIDFromMetadata extracts transaction ID from Stripe metadata.
func parseTransactionIDFromMetadata(metadata map[string]string) (uuid.UUID, error) {
	transactionIDStr, ok := metadata["transaction_id"]
	if !ok {
		return uuid.Nil, errors.New("transaction_id not found in metadata")
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid transaction_id in metadata: %w", err)
	}

	return transactionID, nil
}

// parseRefundMetadata extracts refund and transaction IDs from Stripe refund metadata.
func parseRefundMetadata(
	metadata map[string]string,
) (uuid.UUID, uuid.UUID, error) {
	refundIDStr, ok := metadata["refund_id"]
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New("refund_id not found in metadata")
	}

	refundID, err := uuid.Parse(refundIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid refund_id in metadata: %w", err)
	}

	transactionIDStr, ok := metadata["transaction_id"]
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New("transaction_id not found in metadata")
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid transaction_id in metadata: %w", err)
	}

	return refundID, transactionID, nil
}
