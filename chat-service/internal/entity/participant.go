package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// Participant represents a user or admin participating in a conversation.
type Participant struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	UserID         uuid.UUID
	UserType       constant.UserType
	Role           constant.ParticipantRole
	JoinedAt       time.Time
	LeftAt         *time.Time
	IsActive       bool
}

// NewParticipant creates a new participant with validation.
func NewParticipant(
	conversationID, userID uuid.UUID,
	userType constant.UserType,
	role constant.ParticipantRole,
) (*Participant, error) {
	participant := &Participant{
		ID:             uuid.New(),
		ConversationID: conversationID,
		UserID:         userID,
		UserType:       userType,
		Role:           role,
		JoinedAt:       time.Now(),
		IsActive:       true,
	}

	if err := participant.validate(); err != nil {
		return nil, err
	}

	return participant, nil
}

// Leave marks the participant as having left the conversation.
func (p *Participant) Leave() {
	now := time.Now()
	p.LeftAt = &now
	p.IsActive = false
}

// validate performs business rule validation for participant.
func (p *Participant) validate() error {
	if p.ConversationID == uuid.Nil {
		return errors.New("conversation_id must not be empty")
	}

	if p.UserID == uuid.Nil {
		return errors.New("user_id must not be empty")
	}

	return nil
}
