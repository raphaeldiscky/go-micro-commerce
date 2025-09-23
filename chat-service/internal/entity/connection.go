package entity

import (
	"time"

	"github.com/google/uuid"
)

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

// UpdateHeartbeat updates the last heartbeat timestamp.
func (c *Connection) UpdateHeartbeat() {
	c.LastHeartbeat = time.Now()
}

// Disconnect marks the connection as inactive.
func (c *Connection) Disconnect() {
	c.IsActive = false
}
