package redis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ChannelBuilder provides a fluent interface for building Redis channel names.
type ChannelBuilder struct {
	parts []string
}

// NewChannelBuilder creates a new channel builder.
func NewChannelBuilder() *ChannelBuilder {
	return &ChannelBuilder{
		parts: make([]string, 0),
	}
}

// Service adds a service name to the channel.
func (cb *ChannelBuilder) Service(serviceName string) *ChannelBuilder {
	cb.parts = append(cb.parts, serviceName)
	return cb
}

// Domain adds a domain name to the channel.
func (cb *ChannelBuilder) Domain(domainName string) *ChannelBuilder {
	cb.parts = append(cb.parts, domainName)
	return cb
}

// Entity adds an entity type to the channel.
func (cb *ChannelBuilder) Entity(entityType string) *ChannelBuilder {
	cb.parts = append(cb.parts, entityType)
	return cb
}

// ID adds an entity ID to the channel.
func (cb *ChannelBuilder) ID(id uuid.UUID) *ChannelBuilder {
	cb.parts = append(cb.parts, id.String())
	return cb
}

// IDString adds an entity ID as string to the channel.
func (cb *ChannelBuilder) IDString(id string) *ChannelBuilder {
	cb.parts = append(cb.parts, id)
	return cb
}

// Event adds an event type to the channel.
func (cb *ChannelBuilder) Event(eventType string) *ChannelBuilder {
	cb.parts = append(cb.parts, eventType)
	return cb
}

// Action adds an action type to the channel.
func (cb *ChannelBuilder) Action(actionType string) *ChannelBuilder {
	cb.parts = append(cb.parts, actionType)
	return cb
}

// Custom adds a custom part to the channel.
func (cb *ChannelBuilder) Custom(part string) *ChannelBuilder {
	cb.parts = append(cb.parts, part)
	return cb
}

// Build constructs the final channel name.
func (cb *ChannelBuilder) Build() string {
	return strings.Join(cb.parts, ":")
}

// BuildPattern constructs a pattern for subscribing to multiple channels.
func (cb *ChannelBuilder) BuildPattern() string {
	return cb.Build() + "*"
}

// Common channel patterns for the microservices architecture.

// ChatChannel creates a chat-related channel.
func ChatChannel(conversationID uuid.UUID) string {
	return NewChannelBuilder().
		Service("chat").
		Entity("conversation").
		ID(conversationID).
		Build()
}

// ConversationChannel generates a channel name for a conversation.
// Alias for ChatChannel for backward compatibility.
func ConversationChannel(conversationID uuid.UUID) string {
	return ChatChannel(conversationID)
}

// UserPresenceChannel creates a user presence channel.
func UserPresenceChannel(userID uuid.UUID) string {
	return NewChannelBuilder().
		Service("chat").
		Entity("presence").
		ID(userID).
		Build()
}

// UserChannel generates a channel name for a user.
// Alias for UserPresenceChannel for backward compatibility.
func UserChannel(userID uuid.UUID) string {
	return UserPresenceChannel(userID)
}

// BroadcastChannel creates a broadcast channel for system-wide messages.
func BroadcastChannel(messageType string) string {
	return NewChannelBuilder().
		Service("system").
		Entity("broadcast").
		Custom(messageType).
		Build()
}

// NotificationShardChannel creates a shard-based notification channel.
// Uses consistent hashing to distribute notifications across fixed shards.
func NotificationShardChannel(shardID int) string {
	return NewChannelBuilder().
		Service("notification").
		Entity("shard").
		IDString(strconv.Itoa(shardID)).
		Build()
}

// OrderChannel creates an order-related channel.
func OrderChannel(orderID uuid.UUID, eventType string) string {
	return NewChannelBuilder().
		Service("order").
		Entity("order").
		ID(orderID).
		Event(eventType).
		Build()
}

// AdminChannel creates an admin-specific channel.
func AdminChannel(adminID uuid.UUID, channelType string) string {
	return NewChannelBuilder().
		Service("admin").
		Entity("admin").
		ID(adminID).
		Custom(channelType).
		Build()
}

// ServiceHealthChannel creates a service health monitoring channel.
func ServiceHealthChannel(serviceName string) string {
	return NewChannelBuilder().
		Service("monitoring").
		Entity("health").
		Custom(serviceName).
		Build()
}

// Pattern builders for subscribing to multiple channels.

// AllChatPattern creates a pattern to subscribe to all chat channels.
func AllChatPattern() string {
	return NewChannelBuilder().
		Service("chat").
		BuildPattern()
}

// AllNotificationPattern creates a pattern to subscribe to all notification channels.
func AllNotificationPattern() string {
	return NewChannelBuilder().
		Service("notification").
		BuildPattern()
}

// AllOrderPattern creates a pattern to subscribe to all order channels.
func AllOrderPattern() string {
	return NewChannelBuilder().
		Service("order").
		BuildPattern()
}

// UserChatPattern creates a pattern to subscribe to all chat channels for a specific user.
func UserChatPattern(userID uuid.UUID) string {
	return NewChannelBuilder().
		Service("chat").
		Entity("conversation").
		Custom(fmt.Sprintf("*:%s:*", userID.String())).
		Build()
}
