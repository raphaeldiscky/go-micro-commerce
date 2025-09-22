package websocket

import (
	"context"

	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/pubsub"
)

// MessagePublisher handles publishing different types of chat messages to Redis.
type MessagePublisher interface {
	PublishMessage(
		ctx context.Context,
		conversationID uuid.UUID,
		message *pkgwebsocket.Message,
		excludeUserID *uuid.UUID,
	) error
}

type messagePublisher struct {
	pubSub *pubsub.ChatPubSub
}

// NewMessagePublisher creates a new message publisher.
func NewMessagePublisher(pubSub *pubsub.ChatPubSub) MessagePublisher {
	return &messagePublisher{
		pubSub: pubSub,
	}
}

// PublishMessage publishes a message to Redis based on its type.
func (mp *messagePublisher) PublishMessage(
	ctx context.Context,
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID *uuid.UUID,
) error {
	// Skip messages that don't need cross-instance broadcasting
	if mp.shouldSkipMessage(message.Type) {
		return nil
	}

	payload := mp.createPayload(message)
	crossMsgType := mp.mapMessageType(message.Type)

	return mp.publishByType(ctx, crossMsgType, conversationID, message, payload, excludeUserID)
}

// shouldSkipMessage determines if a message type should be skipped.
func (mp *messagePublisher) shouldSkipMessage(msgType pkgwebsocket.MessageType) bool {
	switch msgType {
	case pkgwebsocket.MessageTypeHeartbeat,
		pkgwebsocket.MessageTypeError,
		pkgwebsocket.MessageTypeSystem:
		return true
	default:
		return false
	}
}

// createPayload creates the payload for cross-instance messages.
func (mp *messagePublisher) createPayload(message *pkgwebsocket.Message) map[string]interface{} {
	return map[string]interface{}{
		"id":        message.ID,
		"type":      message.Type,
		"channel":   message.Channel,
		"sender_id": message.SenderID,
		"content":   message.Content,
		"timestamp": message.Timestamp,
	}
}

// mapMessageType maps WebSocket message types to cross-instance message types.
func (mp *messagePublisher) mapMessageType(
	msgType pkgwebsocket.MessageType,
) pubsub.CrossInstanceMessageType {
	switch msgType {
	case ChatMessageTypeTyping:
		return pubsub.CrossInstanceMessageTypeTyping
	case ChatMessageTypePresence:
		return pubsub.CrossInstanceMessageTypePresence
	case ChatMessageTypeDeliveryReceipt:
		return pubsub.CrossInstanceMessageTypeDeliveryReceipt
	case ChatMessageTypeReadReceipt:
		return pubsub.CrossInstanceMessageTypeReadReceipt
	case pkgwebsocket.MessageTypeHeartbeat,
		pkgwebsocket.MessageTypeError,
		pkgwebsocket.MessageTypeSystem:
		// These message types are handled by shouldSkipMessage and won't reach here
		return pubsub.CrossInstanceMessageTypeChat
	default:
		return pubsub.CrossInstanceMessageTypeChat
	}
}

// publishByType publishes the message using the appropriate method based on type.
func (mp *messagePublisher) publishByType(
	ctx context.Context,
	crossMsgType pubsub.CrossInstanceMessageType,
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	payload map[string]interface{},
	excludeUserID *uuid.UUID,
) error {
	switch crossMsgType {
	case pubsub.CrossInstanceMessageTypeChat:
		return mp.pubSub.PublishChatMessage(ctx, conversationID, payload, excludeUserID)
	case pubsub.CrossInstanceMessageTypeTyping:
		return mp.pubSub.PublishTypingIndicator(ctx, conversationID, payload, excludeUserID)
	case pubsub.CrossInstanceMessageTypePresence:
		return mp.publishPresenceMessage(ctx, message, payload)
	case pubsub.CrossInstanceMessageTypeDeliveryReceipt:
		return mp.pubSub.PublishDeliveryReceipt(ctx, conversationID, payload, excludeUserID)
	case pubsub.CrossInstanceMessageTypeReadReceipt:
		return mp.pubSub.PublishReadReceipt(ctx, conversationID, payload, excludeUserID)
	default:
		return mp.pubSub.PublishChatMessage(ctx, conversationID, payload, excludeUserID)
	}
}

// publishPresenceMessage handles presence message publishing with sender validation.
func (mp *messagePublisher) publishPresenceMessage(
	ctx context.Context,
	message *pkgwebsocket.Message,
	payload map[string]interface{},
) error {
	if message.SenderID != nil {
		return mp.pubSub.PublishPresenceUpdate(ctx, *message.SenderID, payload)
	}

	return nil
}
