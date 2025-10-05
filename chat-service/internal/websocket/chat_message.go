package websocket

import (
	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// Chat-specific message types.
const (
	ChatMessageTypeChat            pkgwebsocket.MessageType = "chat"
	ChatMessageTypeTyping          pkgwebsocket.MessageType = "typing"
	ChatMessageTypePresence        pkgwebsocket.MessageType = "presence"
	ChatMessageTypeDeliveryReceipt pkgwebsocket.MessageType = "delivery_receipt"
	ChatMessageTypeReadReceipt     pkgwebsocket.MessageType = "read_receipt"
	ChatMessageTypeSystem          pkgwebsocket.MessageType = "system"
	ChatMessageTypeError           pkgwebsocket.MessageType = "error"
	ChatMessageTypeHeartbeat       pkgwebsocket.MessageType = "heartbeat"
)

// MessageMetadata represents structured metadata for messages.
type MessageMetadata struct {
	Edited    bool       `json:"edited,omitempty"`
	EditedAt  *int64     `json:"edited_at,omitempty"`
	ReplyToID *uuid.UUID `json:"reply_to_id,omitempty"`
	ThreadID  *uuid.UUID `json:"thread_id,omitempty"`
	Priority  string     `json:"priority,omitempty"`
}

// ChatContent represents the content of a chat message.
type ChatContent struct {
	ConversationID uuid.UUID            `json:"conversation_id"    validate:"required"`
	Text           string               `json:"text"`
	MessageType    constant.MessageType `json:"message_type"`
	Metadata       *MessageMetadata     `json:"metadata,omitempty"`
}

// TypingContent represents typing indicator content.
type TypingContent struct {
	IsTyping bool `json:"is_typing"`
}

// PresenceContent represents presence update content.
type PresenceContent struct {
	UserID uuid.UUID                   `json:"user_id"`
	Status constant.PresenceStatus     `json:"status"`
	Event  constant.WebSocketEventType `json:"event"`
}

// DeliveryReceiptContent represents delivery receipt content.
type DeliveryReceiptContent struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	RecipientID    uuid.UUID `json:"recipient_id"`
	DeliveredAt    int64     `json:"delivered_at"`
}

// ReadReceiptContent represents read receipt content.
type ReadReceiptContent struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	ReaderID       uuid.UUID `json:"reader_id"`
	ReadAt         int64     `json:"read_at"`
}

// NewChatMessage creates a new chat message.
func NewChatMessage(
	conversationID, senderID uuid.UUID,
	text string,
	messageType constant.MessageType,
) (*pkgwebsocket.Message, error) {
	content := ChatContent{
		Text:        text,
		MessageType: messageType,
	}

	msg, err := pkgwebsocket.NewMessage(ChatMessageTypeChat, content)
	if err != nil {
		return nil, err
	}

	// Set chat-specific fields
	conversationChannel := ConversationChannel(conversationID)
	msg.WithChannel(conversationChannel).WithSender(senderID)

	return msg, nil
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(
	content string,
	metadata *MessageMetadata,
) (*pkgwebsocket.Message, error) {
	systemContent := ChatContent{
		Text:        content,
		MessageType: constant.MessageTypeSystem,
		Metadata:    metadata,
	}

	return pkgwebsocket.NewMessage(ChatMessageTypeChat, systemContent)
}

// NewTypingMessage creates a new typing indicator message.
func NewTypingMessage(
	conversationID, senderID uuid.UUID,
	isTyping bool,
) (*pkgwebsocket.Message, error) {
	content := TypingContent{
		IsTyping: isTyping,
	}

	msg, err := pkgwebsocket.NewMessage(ChatMessageTypeTyping, content)
	if err != nil {
		return nil, err
	}

	conversationChannel := ConversationChannel(conversationID)
	msg.WithChannel(conversationChannel).WithSender(senderID)

	return msg, nil
}

// NewPresenceMessage creates a new presence update message.
func NewPresenceMessage(
	userID uuid.UUID,
	status constant.PresenceStatus,
	event constant.WebSocketEventType,
) (*pkgwebsocket.Message, error) {
	content := PresenceContent{
		UserID: userID,
		Status: status,
		Event:  event,
	}

	msg, err := pkgwebsocket.NewMessage(ChatMessageTypePresence, content)
	if err != nil {
		return nil, err
	}

	userChannel := UserChannel(userID)
	msg.WithChannel(userChannel).WithSender(userID)

	return msg, nil
}

// NewDeliveryReceiptMessage creates a new delivery receipt message.
func NewDeliveryReceiptMessage(
	messageID, conversationID, recipientID uuid.UUID,
	deliveredAt int64,
) (*pkgwebsocket.Message, error) {
	content := DeliveryReceiptContent{
		MessageID:      messageID,
		ConversationID: conversationID,
		RecipientID:    recipientID,
		DeliveredAt:    deliveredAt,
	}

	msg, err := pkgwebsocket.NewMessage(ChatMessageTypeDeliveryReceipt, content)
	if err != nil {
		return nil, err
	}

	conversationChannel := ConversationChannel(conversationID)
	msg.WithChannel(conversationChannel).WithSender(recipientID)

	return msg, nil
}

// NewReadReceiptMessage creates a new read receipt message.
func NewReadReceiptMessage(
	messageID, conversationID, readerID uuid.UUID,
	readAt int64,
) (*pkgwebsocket.Message, error) {
	content := ReadReceiptContent{
		MessageID:      messageID,
		ConversationID: conversationID,
		ReaderID:       readerID,
		ReadAt:         readAt,
	}

	msg, err := pkgwebsocket.NewMessage(ChatMessageTypeReadReceipt, content)
	if err != nil {
		return nil, err
	}

	conversationChannel := ConversationChannel(conversationID)
	msg.WithChannel(conversationChannel).WithSender(readerID)

	return msg, nil
}
