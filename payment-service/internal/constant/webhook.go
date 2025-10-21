package constant

import "time"

// Stripe webhook constants.
const (
	// StripeSignatureHeader is the header containing the webhook signature.
	StripeSignatureHeader = "Stripe-Signature"

	// WebhookRequestTimeout is the timeout for processing webhook requests.
	WebhookRequestTimeout = 30 * time.Second
)

// StripeWebhookEventType represents Stripe webhook event types we handle.
type StripeWebhookEventType string

const (
	// StripeEventPaymentIntentSucceeded is triggered when a payment is successfully completed.
	StripeEventPaymentIntentSucceeded StripeWebhookEventType = "payment_intent.succeeded"

	// StripeEventPaymentIntentFailed is triggered when a payment fails.
	StripeEventPaymentIntentFailed StripeWebhookEventType = "payment_intent.failed"

	// StripeEventPaymentIntentCanceled is triggered when a payment is canceled.
	StripeEventPaymentIntentCanceled StripeWebhookEventType = "payment_intent.canceled"

	// StripeEventPaymentIntentRequiresAction is triggered when a payment requires additional action.
	StripeEventPaymentIntentRequiresAction StripeWebhookEventType = "payment_intent.requires_action"

	// StripeEventChargeRefunded is triggered when a charge is refunded.
	StripeEventChargeRefunded StripeWebhookEventType = "charge.refunded"
)
