package dto

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// ConnectionStatsResponse represents WebSocket connection statistics.
type ConnectionStatsResponse struct {
	TotalConnections int `json:"total_connections"`
	UniqueUsers      int `json:"unique_users"`
}

// HealthStatusResponse represents the health status response.
type HealthStatusResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// PresenceUpdateResponse represents a presence update response.
type PresenceUpdateResponse struct {
	UserID  uuid.UUID               `json:"user_id"`
	Status  constant.PresenceStatus `json:"status"`
	Message string                  `json:"message"`
}

// TypingIndicatorResponse represents a typing indicator response.
type TypingIndicatorResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsTyping       bool      `json:"is_typing"`
	Message        string    `json:"message"`
}

// OnlineUsersResponse represents an online users response.
type OnlineUsersResponse struct {
	OnlineUsers []uuid.UUID `json:"online_users"`
	Count       int         `json:"count"`
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
