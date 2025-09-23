// Package dto provides data transfer objects for the chat service.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// CreateConversationRequest represents the request to create a new conversation.
type CreateConversationRequest struct {
	Subject  *string `json:"subject,omitempty" validate:"omitempty,max=255"`
	Priority int     `json:"priority"          validate:"required,min=1,max=4"`
}

// UpdateConversationStatusRequest represents the request to update conversation status.
type UpdateConversationStatusRequest struct {
	Status constant.ConversationStatus `json:"status" validate:"required"`
}

// SetConversationSubjectRequest represents the request to set conversation subject.
type SetConversationSubjectRequest struct {
	Subject string `json:"subject" validate:"required,max=255"`
}

// CreateMessageRequest represents the request to create a new message.
type CreateMessageRequest struct {
	Content     string               `json:"content"            validate:"required,max=1000"`
	MessageType constant.MessageType `json:"message_type"       validate:"required"`
	Metadata    map[string]any       `json:"metadata,omitempty"`
}

// JoinConversationRequest represents the request to join a conversation.
type JoinConversationRequest struct {
	Role constant.ParticipantRole `json:"role" validate:"required"`
}

// ConversationResponse represents the response for conversation operations.
type ConversationResponse struct {
	ID        uuid.UUID                   `json:"id"`
	Status    constant.ConversationStatus `json:"status"`
	Subject   *string                     `json:"subject,omitempty"`
	Priority  int                         `json:"priority"`
	Metadata  map[string]any              `json:"metadata,omitempty"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	EndedAt   *time.Time                  `json:"ended_at,omitempty"`
}

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

// ParticipantResponse represents the response for participant operations.
type ParticipantResponse struct {
	ID             uuid.UUID                `json:"id"`
	ConversationID uuid.UUID                `json:"conversation_id"`
	UserID         uuid.UUID                `json:"user_id"`
	UserType       constant.UserType        `json:"user_type"`
	Role           constant.ParticipantRole `json:"role"`
	JoinedAt       time.Time                `json:"joined_at"`
	LeftAt         *time.Time               `json:"left_at,omitempty"`
	IsActive       bool                     `json:"is_active"`
}

// ConnectionResponse represents the response for connection operations.
type ConnectionResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ConnectionID  string    `json:"connection_id"`
	SocketID      string    `json:"socket_id"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	IsActive      bool      `json:"is_active"`
}

// ConversationListResponse represents a paginated list of conversations.
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PerPage       int                    `json:"per_page"`
	TotalPages    int                    `json:"total_pages"`
}

// MessageListResponse represents a paginated list of messages.
type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

// ChatStatsResponse represents chat service statistics.
type ChatStatsResponse struct {
	TotalConversations   int64 `json:"total_conversations"`
	ActiveConversations  int64 `json:"active_conversations"`
	WaitingConversations int64 `json:"waiting_conversations"`
	TotalMessages        int64 `json:"total_messages"`
	OnlineUsers          int64 `json:"online_users"`
	OnlineAdmins         int64 `json:"online_admins"`
}

// UpdatePresenceRequest represents the request to update user presence.
type UpdatePresenceRequest struct {
	Status constant.PresenceStatus `json:"status" validate:"required"`
}

// TypingIndicatorRequest represents the request to send typing indicator.
type TypingIndicatorRequest struct {
	IsTyping bool `json:"is_typing"`
}
