package stripe

import (
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
	paymentID uuid.UUID,
) (*dto.PaymentGatewayResponse, error) {
	// Convert amount from cents to decimal
	amount := decimal.NewFromInt(pi.Amount).Div(decimal.NewFromInt(multiplyAmount))

	// Map Stripe status to our internal status
	status := mapStripeStatusToGatewayStatus(string(pi.Status))

	// Set client_secret for frontend to confirm payment with Stripe.js
	clientSecret := pi.ClientSecret

	// Build Stripe-specific metadata
	gatewayMetadata := map[string]any{
		"payment_intent_id": pi.ID,
		"status":            string(pi.Status),
		"client_secret":     pi.ClientSecret,
	}

	// Add payment method if available
	if pi.PaymentMethod != nil {
		gatewayMetadata["payment_method_id"] = pi.PaymentMethod.ID
	}

	// Add customer ID if available
	if pi.Customer != nil {
		gatewayMetadata["customer_id"] = pi.Customer.ID
	}

	// Add charge ID if available (after successful payment)
	if pi.LatestCharge != nil {
		gatewayMetadata["charge_id"] = pi.LatestCharge.ID
	}

	response := &dto.PaymentGatewayResponse{
		PaymentID:       paymentID,
		PaymentIntentID: pi.ID,
		Status:          status,
		Amount:          amount,
		Currency:        string(pi.Currency),
		ProcessedAt:     time.Unix(pi.Created, 0),
		ClientSecret:    &clientSecret,   // For stripe.confirmCardPayment() on frontend
		GatewayResponse: gatewayMetadata, // Stripe-specific metadata
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
	refundID, paymentID uuid.UUID,
) (*dto.RefundResponse, error) {
	// Convert amount from cents to decimal
	amount := decimal.NewFromInt(r.Amount).Div(decimal.NewFromInt(multiplyAmount))

	// Map Stripe refund status to our internal status
	status := mapStripeRefundStatus(string(r.Status))

	response := &dto.RefundResponse{
		RefundID:       refundID,
		PaymentID:      paymentID,
		StripeRefundID: r.ID,
		Status:         status,
		Amount:         amount,
		Currency:       string(r.Currency),
		ProcessedAt:    time.Unix(r.Created, 0),
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
