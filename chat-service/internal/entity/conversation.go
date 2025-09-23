// Package entity defines the chat entities and their validation logic.
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// Conversation represents a chat conversation between users and admins.
type Conversation struct {
	ID        uuid.UUID
	Status    constant.ConversationStatus
	Subject   *string
	Priority  int
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	EndedAt   *time.Time
}

// NewConversation creates a new conversation with validation.
func NewConversation(subject *string, priority int) (*Conversation, error) {
	now := time.Now()
	conversation := &Conversation{
		ID:        uuid.New(),
		Status:    constant.ConversationStatusWaiting,
		Subject:   subject,
		Priority:  priority,
		Metadata:  make(map[string]any),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := conversation.validate(); err != nil {
		return nil, err
	}

	return conversation, nil
}

// UpdateStatus updates the conversation status with validation.
func (c *Conversation) UpdateStatus(status constant.ConversationStatus) error {
	c.Status = status
	c.UpdatedAt = time.Now()

	// Set end timestamp for ended status
	if status == constant.ConversationStatusEnded && c.EndedAt == nil {
		now := time.Now()
		c.EndedAt = &now
	}

	return c.validate()
}

// SetSubject sets the conversation subject.
func (c *Conversation) SetSubject(subject string) {
	c.Subject = &subject
	c.UpdatedAt = time.Now()
}

// IsActive checks if conversation is active.
func (c *Conversation) IsActive() bool {
	return c.Status == constant.ConversationStatusActive
}

// IsWaiting checks if conversation is waiting for admin.
func (c *Conversation) IsWaiting() bool {
	return c.Status == constant.ConversationStatusWaiting
}

// IsEnded checks if conversation has ended.
func (c *Conversation) IsEnded() bool {
	return c.Status == constant.ConversationStatusEnded
}

// validate performs business rule validation for conversation.
func (c *Conversation) validate() error {
	if c.Priority < 1 || c.Priority > 4 {
		return errors.New("priority must be between 1 and 4")
	}

	if c.CreatedAt.After(c.UpdatedAt) {
		return errors.New("created_at must be before or equal to updated_at")
	}

	return nil
}
