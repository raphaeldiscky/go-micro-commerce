package kafka

const (
	// CheckoutSessionLifecycleTopic is the topic for checkout session lifecycle events.
	CheckoutSessionLifecycleTopic = "checkout_session.lifecycle"
)

const (
	// CheckoutSessionCreatedEventType is the event type for checkout session created events.
	CheckoutSessionCreatedEventType = "CheckoutSessionCreated"
	// CheckoutSessionOrderPlacedEventType is the event type for checkout session placed order events.
	CheckoutSessionOrderPlacedEventType = "CheckoutSessionOrderPlaced"
	// CheckoutSessionCanceledEventType is the event type for checkout session deleted events.
	CheckoutSessionCanceledEventType = "CheckoutSessionCanceled"
)
