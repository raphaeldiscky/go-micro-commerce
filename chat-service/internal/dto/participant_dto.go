package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

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
