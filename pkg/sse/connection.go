// Package sse provides utilities for SSE connections.
package sse

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Connection represents an SSE connection to a client.
type Connection struct {
	id            uuid.UUID
	userID        uuid.UUID
	echoCtx       echo.Context
	send          chan *Message
	done          chan struct{}
	mu            sync.RWMutex
	isActive      bool
	lastHeartbeat time.Time
	logger        logger.Logger
}

// NewConnection creates a new SSE connection.
func NewConnection(
	userID uuid.UUID,
	echoCtx echo.Context,
	appLogger logger.Logger,
) *Connection {
	return &Connection{
		id:            uuid.New(),
		userID:        userID,
		echoCtx:       echoCtx,
		send:          make(chan *Message, constant.SSEMessageBufferSize),
		done:          make(chan struct{}),
		isActive:      true,
		lastHeartbeat: time.Now(),
		logger:        appLogger,
	}
}

// ID returns the connection ID.
func (c *Connection) ID() uuid.UUID {
	return c.id
}

// UserID returns the user ID.
func (c *Connection) UserID() uuid.UUID {
	return c.userID
}

// IsActive returns whether the connection is active.
func (c *Connection) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isActive
}

// GetLastHeartbeat returns the last heartbeat time.
func (c *Connection) GetLastHeartbeat() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastHeartbeat
}

// Send sends a message to the client.
func (c *Connection) Send(message *Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isActive {
		return ErrConnectionClosed
	}

	select {
	case c.send <- message:
		return nil
	case <-c.done:
		return ErrConnectionClosed
	default:
		return errors.New("send buffer full")
	}
}

// WritePump writes messages to the SSE stream.
func (c *Connection) WritePump(ctx context.Context) {
	heartbeatTicker := time.NewTicker(constant.SSEHeartbeatTicker)
	defer heartbeatTicker.Stop()

	// Set SSE headers
	resp := c.echoCtx.Response()
	resp.Header().Set(echo.HeaderContentType, "text/event-stream")
	resp.Header().Set(echo.HeaderCacheControl, "no-cache")
	resp.Header().Set(echo.HeaderConnection, "keep-alive")
	resp.WriteHeader(http.StatusOK)

	// Send initial connection established message
	initialMsg := "event: connected\ndata: {\"status\":\"connected\"}\n\n"
	if _, err := resp.Write([]byte(initialMsg)); err != nil {
		c.logger.Error("Failed to write initial message",
			"connection_id", c.id,
			"error", err)

		if err = c.Close(); err != nil {
			c.logger.Error("Failed to close connection", "connection_id", c.id, "error", err)
		}

		return
	}

	resp.Flush()

	for {
		select {
		case <-ctx.Done():
			if err := c.Close(); err != nil {
				c.logger.Error("Failed to close connection", "connection_id", c.id, "error", err)
			}

			return

		case <-c.done:
			return

		case message := <-c.send:
			// Format and send SSE message
			formattedMsg := message.Format()

			if _, err := resp.Write([]byte(formattedMsg)); err != nil {
				c.logger.Error("Failed to write message",
					"connection_id", c.id,
					"user_id", c.userID,
					"error", err)

				if err = c.Close(); err != nil {
					c.logger.Error(
						"Failed to close connection",
						"connection_id",
						c.id,
						"error",
						err,
					)
				}

				return
			}

			resp.Flush()

			c.updateHeartbeat()

		case <-heartbeatTicker.C:
			// Send heartbeat comment to keep connection alive
			heartbeat := ": heartbeat\n\n"

			if _, err := resp.Write([]byte(heartbeat)); err != nil {
				c.logger.Error("Failed to write heartbeat",
					"connection_id", c.id,
					"user_id", c.userID,
					"error", err)

				if err = c.Close(); err != nil {
					c.logger.Error(
						"Failed to close connection",
						"connection_id",
						c.id,
						"error",
						err,
					)
				}

				return
			}

			resp.Flush()

			c.updateHeartbeat()
		}
	}
}

// Close closes the connection.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isActive {
		return nil
	}

	c.isActive = false
	close(c.done)

	c.logger.Info("SSE connection closed",
		"connection_id", c.id,
		"user_id", c.userID)

	return nil
}

// updateHeartbeat updates the last heartbeat time.
func (c *Connection) updateHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastHeartbeat = time.Now()
}
