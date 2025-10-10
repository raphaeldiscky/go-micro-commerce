// Package pubsub provides Redis pub/sub functionality for cross-instance chat messaging.
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	redispkg "github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

// ChatPubSub provides chat-specific pub/sub functionality.
type ChatPubSub struct {
	publisher       redispkg.Publisher
	subscriber      redispkg.Subscriber
	logger          logger.Logger
	instanceID      string
	messageHandlers map[CrossInstanceMessageType][]MessageHandler
}

// MessageHandler defines the function signature for handling cross-instance messages.
type MessageHandler func(ctx context.Context, message *CrossInstanceMessage) error

// CrossInstanceMessageType represents the type of cross-instance message.
type CrossInstanceMessageType string

const (
	// CrossInstanceMessageTypeChat represents a chat message.
	CrossInstanceMessageTypeChat CrossInstanceMessageType = "chat"
	// CrossInstanceMessageTypeTyping represents a typing indicator.
	CrossInstanceMessageTypeTyping CrossInstanceMessageType = "typing"
	// CrossInstanceMessageTypePresence represents a presence update.
	CrossInstanceMessageTypePresence CrossInstanceMessageType = "presence"
	// CrossInstanceMessageTypeDeliveryReceipt represents a delivery receipt.
	CrossInstanceMessageTypeDeliveryReceipt CrossInstanceMessageType = "delivery_receipt"
	// CrossInstanceMessageTypeReadReceipt represents a read receipt.
	CrossInstanceMessageTypeReadReceipt CrossInstanceMessageType = "read_receipt"
)

// CrossInstanceMessage represents a message sent between chat service instances.
type CrossInstanceMessage struct {
	// SourceInstanceID identifies which instance sent this message.
	SourceInstanceID string `json:"source_instance_id"`
	// MessageType indicates the type of message.
	MessageType CrossInstanceMessageType `json:"message_type"`
	// ConversationID is the target conversation (for conversation-specific messages).
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	// UserID is the target user (for user-specific messages like presence).
	UserID *uuid.UUID `json:"user_id,omitempty"`
	// ExcludeUserID is used to exclude a specific user from receiving the message.
	ExcludeUserID *uuid.UUID `json:"exclude_user_id,omitempty"`
	// Payload contains the actual message content.
	Payload json.RawMessage `json:"payload"`
	// Timestamp when the message was created.
	Timestamp time.Time `json:"timestamp"`
}

// NewChatPubSub creates a new chat pub/sub service.
func NewChatPubSub(
	publisher redispkg.Publisher,
	subscriber redispkg.Subscriber,
	logger logger.Logger,
) *ChatPubSub {
	instanceID := uuid.New().String()

	return &ChatPubSub{
		publisher:       publisher,
		subscriber:      subscriber,
		logger:          logger,
		instanceID:      instanceID,
		messageHandlers: make(map[CrossInstanceMessageType][]MessageHandler),
	}
}

// RegisterHandler registers a handler for a specific message type.
// Multiple handlers can be registered for the same message type and all will be called.
func (c *ChatPubSub) RegisterHandler(msgType CrossInstanceMessageType, handler MessageHandler) {
	c.messageHandlers[msgType] = append(c.messageHandlers[msgType], handler)
	c.logger.Infof("Registered handler for message type: %s (total handlers: %d)",
		msgType, len(c.messageHandlers[msgType]))
}

// PublishChatMessage publishes a chat message to other instances.
func (c *ChatPubSub) PublishChatMessage(
	ctx context.Context,
	conversationID uuid.UUID,
	payload any,
	excludeUserID *uuid.UUID,
) error {
	return c.publishMessage(
		ctx,
		CrossInstanceMessageTypeChat,
		&conversationID,
		nil,
		excludeUserID,
		payload,
	)
}

// PublishTypingIndicator publishes a typing indicator to other instances.
func (c *ChatPubSub) PublishTypingIndicator(
	ctx context.Context,
	conversationID uuid.UUID,
	payload any,
	excludeUserID *uuid.UUID,
) error {
	return c.publishMessage(
		ctx,
		CrossInstanceMessageTypeTyping,
		&conversationID,
		nil,
		excludeUserID,
		payload,
	)
}

// PublishPresenceUpdate publishes a presence update to other instances.
func (c *ChatPubSub) PublishPresenceUpdate(
	ctx context.Context,
	userID uuid.UUID,
	payload any,
) error {
	return c.publishMessage(ctx, CrossInstanceMessageTypePresence, nil, &userID, nil, payload)
}

// PublishDeliveryReceipt publishes a delivery receipt to other instances.
func (c *ChatPubSub) PublishDeliveryReceipt(
	ctx context.Context,
	conversationID uuid.UUID,
	payload any,
	excludeUserID *uuid.UUID,
) error {
	return c.publishMessage(
		ctx,
		CrossInstanceMessageTypeDeliveryReceipt,
		&conversationID,
		nil,
		excludeUserID,
		payload,
	)
}

// PublishReadReceipt publishes a read receipt to other instances.
func (c *ChatPubSub) PublishReadReceipt(
	ctx context.Context,
	conversationID uuid.UUID,
	payload any,
	excludeUserID *uuid.UUID,
) error {
	return c.publishMessage(
		ctx,
		CrossInstanceMessageTypeReadReceipt,
		&conversationID,
		nil,
		excludeUserID,
		payload,
	)
}

// publishMessage publishes a cross-instance message to Redis.
func (c *ChatPubSub) publishMessage(
	ctx context.Context,
	msgType CrossInstanceMessageType,
	conversationID *uuid.UUID,
	userID *uuid.UUID,
	excludeUserID *uuid.UUID,
	payload any,
) error {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create cross-instance message
	crossMsg := &CrossInstanceMessage{
		SourceInstanceID: c.instanceID,
		MessageType:      msgType,
		ConversationID:   conversationID,
		UserID:           userID,
		ExcludeUserID:    excludeUserID,
		Payload:          payloadBytes,
		Timestamp:        time.Now(),
	}

	// Create Redis message with metadata
	metadata := redispkg.NewMessageMetadata("chat-service")

	redisMsg, err := redispkg.NewMessage(metadata, crossMsg)
	if err != nil {
		return fmt.Errorf("failed to create Redis message: %w", err)
	}

	// Determine channel based on message type
	channel := c.getChannelForMessage(msgType, conversationID, userID)

	// Publish to Redis
	if err = c.publisher.Publish(ctx, channel, redisMsg); err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}

	c.logger.Debug("Published cross-instance message",
		"message_type", msgType,
		"channel", channel,
		"instance_id", c.instanceID)

	return nil
}

// getChannelForMessage determines the appropriate Redis channel for a message.
func (c *ChatPubSub) getChannelForMessage(
	msgType CrossInstanceMessageType,
	conversationID *uuid.UUID,
	userID *uuid.UUID,
) string {
	switch msgType {
	case CrossInstanceMessageTypePresence:
		if userID != nil {
			return redispkg.UserPresenceChannel(*userID)
		}

		return redispkg.BroadcastChannel("presence")

	case CrossInstanceMessageTypeChat,
		CrossInstanceMessageTypeTyping,
		CrossInstanceMessageTypeDeliveryReceipt,
		CrossInstanceMessageTypeReadReceipt:
		if conversationID != nil {
			return redispkg.ChatChannel(*conversationID)
		}

		return redispkg.BroadcastChannel("chat")

	default:
		return redispkg.BroadcastChannel("chat")
	}
}

// StartSubscriber starts subscribing to Redis channels for cross-instance messages.
func (c *ChatPubSub) StartSubscriber(ctx context.Context) error {
	// Subscribe to all chat-related patterns
	pattern := redispkg.AllChatPattern() // "chat:*"

	return c.subscriber.SubscribePattern(ctx, c.handleRedisMessage, pattern)
}

// handleRedisMessage handles incoming Redis messages from other instances.
func (c *ChatPubSub) handleRedisMessage(ctx context.Context, redisMsg *redispkg.Message) error {
	// Parse the cross-instance message
	var crossMsg CrossInstanceMessage
	if err := redisMsg.UnmarshalPayload(&crossMsg); err != nil {
		return fmt.Errorf("failed to unmarshal cross-instance message: %w", err)
	}

	// Skip messages from our own instance to avoid loops
	if crossMsg.SourceInstanceID == c.instanceID {
		return nil
	}

	c.logger.Debug("Received cross-instance message",
		"message_type", crossMsg.MessageType,
		"source_instance", crossMsg.SourceInstanceID,
		"our_instance", c.instanceID)

	// Find and call all registered handlers for this message type
	handlers, exists := c.messageHandlers[crossMsg.MessageType]
	if !exists || len(handlers) == 0 {
		c.logger.Warn("No handler registered for message type", "type", crossMsg.MessageType)
		return nil
	}

	// Call all handlers for this message type
	var errs []error

	for i, handler := range handlers {
		if err := handler(ctx, &crossMsg); err != nil {
			c.logger.Error("Handler failed for message type",
				"type", crossMsg.MessageType,
				"handler_index", i,
				"error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"one or more handlers failed for message type %s: %d errors",
			crossMsg.MessageType,
			len(errs),
		)
	}

	return nil
}

// Shutdown gracefully shuts down the pub/sub service.
func (c *ChatPubSub) Shutdown() error {
	c.logger.Info("Shutting down chat pub/sub service...")

	var errs []error

	if err := c.subscriber.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close subscriber: %w", err))
	}

	if err := c.publisher.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close publisher: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	c.logger.Info("Chat pub/sub service shut down successfully")

	return nil
}
