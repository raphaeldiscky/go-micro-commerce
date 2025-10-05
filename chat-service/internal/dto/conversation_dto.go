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

// JoinConversationRequest represents the request to join a conversation.
type JoinConversationRequest struct {
	Role constant.ParticipantRole `json:"role" validate:"required"`
}

// ConversationResponse represents the response for conversation operations.
type ConversationResponse struct {
	ID               uuid.UUID                   `json:"id"`
	Status           constant.ConversationStatus `json:"status"`
	Subject          *string                     `json:"subject,omitempty"`
	Priority         int                         `json:"priority"`
	ParticipantCount int                         `json:"participant_count"`
	Metadata         map[string]any              `json:"metadata,omitempty"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
	EndedAt          *time.Time                  `json:"ended_at,omitempty"`
}

// ConversationListResponse represents a paginated list of conversations.
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PerPage       int                    `json:"per_page"`
	TotalPages    int                    `json:"total_pages"`
}
