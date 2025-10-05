package constant

import "time"

const (
	// SubscriptionChannelBufferSize is the buffer size for subscription channels.
	SubscriptionChannelBufferSize = 10

	// SubscriptionMessageChannelBufferSize is the buffer size for message channels in subscriptions.
	SubscriptionMessageChannelBufferSize = 100

	// GraphQLKeepAlivePingInterval is the keep alive ping interval for GraphQL WebSocket subscriptions.
	GraphQLKeepAlivePingInterval = 10 * time.Second
)
