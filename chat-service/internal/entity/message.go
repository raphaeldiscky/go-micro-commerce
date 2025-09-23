package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// Message represents a chat message within a conversation.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       *uuid.UUID // NULL for system messages
	Content        string
	MessageType    constant.MessageType
	Metadata       map[string]any
	IsSystem       bool
	CreatedAt      time.Time
}

// NewMessage creates a new message with validation.
func NewMessage(
	conversationID uuid.UUID,
	senderID uuid.UUID,
	content string,
	messageType constant.MessageType,
) (*Message, error) {
	message := &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       &senderID,
		Content:        content,
		MessageType:    messageType,
		Metadata:       make(map[string]any),
		IsSystem:       false,
		CreatedAt:      time.Now(),
	}

	if err := message.validate(); err != nil {
		return nil, err
	}

	return message, nil
}

// validate performs business rule validation for message.
func (m *Message) validate() error {
	if m.ConversationID == uuid.Nil {
		return errors.New("conversation_id must not be empty")
	}

	if !m.IsSystem && m.SenderID == nil {
		return errors.New("sender_id must not be empty for non-system messages")
	}

	if m.IsSystem && m.SenderID != nil {
		return errors.New("sender_id must be empty for system messages")
	}

	if m.Content == "" {
		return errors.New("content must not be empty")
	}

	return nil
}
