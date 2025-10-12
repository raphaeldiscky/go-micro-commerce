package constant

import "time"

const (
	// DefaultEntityNotificationLimit is the default limit for fetching notifications in entity resolvers.
	DefaultEntityNotificationLimit = 20
	// SubscriptionChannelBufferSize is the buffer size for GraphQL subscription channels.
	SubscriptionChannelBufferSize = 100
	// SubscriptionKeepAlivePingInterval is the keep alive ping interval for GraphQL subscriptions.
	SubscriptionKeepAlivePingInterval = 10 * time.Second
)
