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

// Connection represents an active WebSocket connection.
type Connection struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ConnectionID  string
	SocketID      string
	UserAgent     *string
	IPAddress     *string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	IsActive      bool
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

// UpdateHeartbeat updates the last heartbeat timestamp.
func (c *Connection) UpdateHeartbeat() {
	c.LastHeartbeat = time.Now()
}

// Disconnect marks the connection as inactive.
func (c *Connection) Disconnect() {
	c.IsActive = false
}
