package sse

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message represents an SSE message to be sent to clients.
type Message struct {
	ID        uuid.UUID       `json:"id"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Retry     *int            `json:"retry,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewMessage creates a new SSE message.
func NewMessage(event string, data any) (*Message, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        uuid.New(),
		Event:     event,
		Data:      dataJSON,
		CreatedAt: time.Now(),
	}, nil
}

// Format formats the message according to SSE protocol.
func (m *Message) Format() string {
	formatted := ""

	if m.ID != uuid.Nil {
		formatted += "id: " + m.ID.String() + "\n"
	}

	if m.Event != "" {
		formatted += "event: " + m.Event + "\n"
	}

	if len(m.Data) > 0 {
		formatted += "data: " + string(m.Data) + "\n"
	}

	if m.Retry != nil {
		formatted += "retry: " + string(rune(*m.Retry)) + "\n"
	}

	formatted += "\n"

	return formatted
}
