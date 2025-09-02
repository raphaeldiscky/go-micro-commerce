// Package topic contains event topics.
package topic

const (
	// OrderDLQTopic is the dead-letter queue topic for failed order events.
	OrderDLQTopic = "order.dlq"
	// OrderLifecycleTopic is the topic for order lifecycle events.
	OrderLifecycleTopic = "order.lifecycle"
)
