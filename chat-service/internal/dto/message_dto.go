package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// MessageResponse represents the response for message operations.
type MessageResponse struct {
	ID             uuid.UUID            `json:"id"`
	ConversationID uuid.UUID            `json:"conversation_id"`
	SenderID       *uuid.UUID           `json:"sender_id,omitempty"`
	Content        string               `json:"content"`
	MessageType    constant.MessageType `json:"message_type"`
	Metadata       map[string]any       `json:"metadata,omitempty"`
	IsSystem       bool                 `json:"is_system"`
	CreatedAt      time.Time            `json:"created_at"`
}

// CreateMessageRequest represents the request to create a new message.
type CreateMessageRequest struct {
	Content     string               `json:"content"            validate:"required,max=1000"`
	MessageType constant.MessageType `json:"message_type"       validate:"required"`
	Metadata    map[string]any       `json:"metadata,omitempty"`
}

// MessageListResponse represents a paginated list of messages.
type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}
