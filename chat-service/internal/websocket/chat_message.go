package websocket

import (
	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// Chat-specific message types.
const (
	ChatMessageTypeChat     pkgwebsocket.MessageType = "chat"
	ChatMessageTypeTyping   pkgwebsocket.MessageType = "typing"
	ChatMessageTypePresence pkgwebsocket.MessageType = "presence"
)

// ChatContent represents the content of a chat message.
type ChatContent struct {
	Text        string               `json:"text"`
	MessageType constant.MessageType `json:"message_type"`
	Metadata    map[string]any       `json:"metadata,omitempty"`
}

// TypingContent represents typing indicator content.
type TypingContent struct {
	IsTyping bool `json:"is_typing"`
}

// PresenceContent represents presence update content.
type PresenceContent struct {
	UserID uuid.UUID                   `json:"user_id"`
	Status string                      `json:"status"`
	Event  constant.WebSocketEventType `json:"event"`
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
	conversationChannel := "conversation:" + conversationID.String()
	msg.WithChannel(conversationChannel).WithSender(senderID)

	return msg, nil
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(
	content string,
	metadata map[string]any,
) (*pkgwebsocket.Message, error) {
	systemContent := ChatContent{
		Text:        content,
		MessageType: constant.MessageTypeSystem,
		Metadata:    metadata,
	}

	return pkgwebsocket.NewMessage(ChatMessageTypeChat, systemContent)
}
