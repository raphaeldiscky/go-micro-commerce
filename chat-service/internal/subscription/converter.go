// Package subscription provides GraphQL subscription infrastructure for bridging webSocket events to GraphQL subscriptions over graphql-transport-ws protocol.
package subscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// EventConverter converts WebSocket messages to GraphQL subscription events.
type EventConverter struct {
	logger logger.Logger
}

// NewEventConverter creates a new event converter.
func NewEventConverter(logger logger.Logger) *EventConverter {
	return &EventConverter{
		logger: logger,
	}
}

// ToConversationEvent converts a WebSocket message to a GraphQL ConversationEvent.
func (c *EventConverter) ToConversationEvent(
	msg *pkgwebsocket.Message,
) (graph.ConversationEvent, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	var err error

	switch msg.Type {
	case websocket.ChatMessageTypeChat:
		return c.convertToNewMessage(msg)
	case websocket.ChatMessageTypeTyping:
		return c.convertToTypingIndicator(msg)
	case websocket.ChatMessageTypeDeliveryReceipt:
		return c.convertToDeliveryReceipt(msg)
	case websocket.ChatMessageTypeReadReceipt:
		return c.convertToReadReceipt(msg)
	case websocket.ChatMessageTypeHeartbeat,
		websocket.ChatMessageTypeError,
		websocket.ChatMessageTypeSystem:
		c.logger.Warnf("unsupported message type: %s", msg.Type)
		return nil, err
	default:
		return nil, err
	}
}

// ToUserEvent converts a WebSocket message to a GraphQL UserEvent.
func (c *EventConverter) ToUserEvent(msg *pkgwebsocket.Message) (graph.UserEvent, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	var err error

	switch msg.Type {
	case websocket.ChatMessageTypePresence:
		return c.convertToPresenceUpdate(msg)
	case websocket.ChatMessageTypeChat, websocket.ChatMessageTypeTyping,
		websocket.ChatMessageTypeDeliveryReceipt, websocket.ChatMessageTypeReadReceipt,
		websocket.ChatMessageTypeHeartbeat, websocket.ChatMessageTypeError,
		websocket.ChatMessageTypeSystem:
		c.logger.Warnf("unsupported message type: %s", msg.Type)
		return nil, err
	default:
		return nil, err
	}
}

// convertToNewMessage converts a chat message to NewMessage GraphQL type.
func (c *EventConverter) convertToNewMessage(msg *pkgwebsocket.Message) (*graph.NewMessage, error) {
	var content websocket.ChatContent
	if err := c.unmarshalContent(msg.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat content: %w", err)
	}

	// Extract conversation ID from channel (format: "conversation:{uuid}")
	var conversationID uuid.UUID

	if msg.Channel != nil {
		uuidStr := extractUUIDFromChannel(*msg.Channel)

		parsedUUID, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation UUID: %w", err)
		}

		conversationID = parsedUUID
	}

	var senderID uuid.UUID
	if msg.SenderID != nil {
		senderID = *msg.SenderID
	}

	return &graph.NewMessage{
		ID:             msg.ID,
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content.Text,
		MessageType:    content.MessageType,
		IsSystem:       content.MessageType == constant.MessageTypeSystem,
		CreatedAt:      msg.Timestamp,
	}, nil
}

// convertToTypingIndicator converts a typing message to TypingIndicator GraphQL type.
func (c *EventConverter) convertToTypingIndicator(
	msg *pkgwebsocket.Message,
) (*graph.TypingIndicator, error) {
	var content websocket.TypingContent
	if err := c.unmarshalContent(msg.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal typing content: %w", err)
	}

	var conversationID uuid.UUID

	if msg.Channel != nil {
		uuidStr := extractUUIDFromChannel(*msg.Channel)

		parsedUUID, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation UUID: %w", err)
		}

		conversationID = parsedUUID
	}

	var userID uuid.UUID
	if msg.SenderID != nil {
		userID = *msg.SenderID
	}

	return &graph.TypingIndicator{
		UserID:         userID,
		ConversationID: conversationID,
		IsTyping:       content.IsTyping,
		Timestamp:      msg.Timestamp,
	}, nil
}

// convertToDeliveryReceipt converts a delivery receipt message to DeliveryReceipt GraphQL type.
func (c *EventConverter) convertToDeliveryReceipt(
	msg *pkgwebsocket.Message,
) (*graph.DeliveryReceipt, error) {
	var content websocket.DeliveryReceiptContent
	if err := c.unmarshalContent(msg.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delivery receipt content: %w", err)
	}

	return &graph.DeliveryReceipt{
		MessageID:      content.MessageID,
		ConversationID: content.ConversationID,
		RecipientID:    content.RecipientID,
		DeliveredAt:    time.Unix(content.DeliveredAt, 0),
	}, nil
}

// convertToReadReceipt converts a read receipt message to ReadReceipt GraphQL type.
func (c *EventConverter) convertToReadReceipt(
	msg *pkgwebsocket.Message,
) (*graph.ReadReceipt, error) {
	var content websocket.ReadReceiptContent
	if err := c.unmarshalContent(msg.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal read receipt content: %w", err)
	}

	return &graph.ReadReceipt{
		MessageID:      content.MessageID,
		ConversationID: content.ConversationID,
		ReaderID:       content.ReaderID,
		ReadAt:         time.Unix(content.ReadAt, 0),
	}, nil
}

// convertToPresenceUpdate converts a presence message to PresenceUpdate GraphQL type.
func (c *EventConverter) convertToPresenceUpdate(
	msg *pkgwebsocket.Message,
) (*graph.PresenceUpdate, error) {
	var content websocket.PresenceContent
	if err := c.unmarshalContent(msg.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal presence content: %w", err)
	}

	// Status is already the same type - no conversion needed
	status := content.Status

	// LastSeen is optional
	var lastSeen *time.Time
	if !msg.Timestamp.IsZero() {
		lastSeen = &msg.Timestamp
	}

	return &graph.PresenceUpdate{
		UserID:   content.UserID,
		Status:   status,
		LastSeen: lastSeen,
	}, nil
}

// unmarshalContent unmarshals message content to the target struct.
func (c *EventConverter) unmarshalContent(content any, target any) error {
	// Content might be already unmarshaled map or JSON bytes
	switch v := content.(type) {
	case map[string]any:
		// Re-marshal and unmarshal to convert map to struct
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, target)
	case []byte:
		return json.Unmarshal(v, target)
	case string:
		return json.Unmarshal([]byte(v), target)
	default:
		// Try marshaling and unmarshaling
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, target)
	}
}

// extractUUIDFromChannel extracts UUID from channel name (e.g., "conversation:uuid" -> "uuid").
func extractUUIDFromChannel(channel string) string {
	// Channel format: "conversation:{uuid}" or "user:{uuid}"
	// Extract the UUID part after ":"
	if len(channel) > 0 {
		for i, ch := range channel {
			if ch == ':' && i+1 < len(channel) {
				return channel[i+1:]
			}
		}
	}

	return channel
}
