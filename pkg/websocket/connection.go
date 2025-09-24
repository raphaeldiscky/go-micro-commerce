package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Connection represents a universal WebSocket connection interface.
type Connection interface {
	// ID returns the unique connection ID
	ID() uuid.UUID
	// UserID returns the user ID associated with this connection
	UserID() uuid.UUID
	// IsActive returns whether the connection is active
	IsActive() bool
	// Send sends a message to the connection
	Send(message *Message) error
	// Close closes the connection
	Close() error
	// UpdateHeartbeat updates the last heartbeat timestamp
	UpdateHeartbeat()
	// GetLastHeartbeat returns the last heartbeat timestamp
	GetLastHeartbeat() time.Time
}

// ConnectionHandler defines the interface for handling connection events.
type ConnectionHandler interface {
	// OnConnect is called when a connection is established
	OnConnect(conn Connection) error
	// OnDisconnect is called when a connection is closed
	OnDisconnect(conn Connection)
	// OnMessage is called when a message is received
	OnMessage(conn Connection, message *Message) error
	// OnError is called when an error occurs
	OnError(conn Connection, err error)
}

// ConnectionConfig holds configuration for WebSocket connections.
type ConnectionConfig struct {
	ReadBufferSize  int           // Size of the read buffer
	WriteBufferSize int           // Size of the write buffer
	MaxMessageSize  int64         // Maximum message size in bytes
	PongWait        time.Duration // Time allowed to read the next pong message
	GracePeriod     time.Duration // Grace period before closing the connection
	PingPeriod      time.Duration // Send pings to peer with this period
	WriteWait       time.Duration // Time allowed to write a message
	SendBufferSize  int           // Size of the send channel buffer
}

// BaseConnection provides a base implementation of the Connection interface.
type BaseConnection struct {
	id            uuid.UUID
	userID        uuid.UUID
	conn          *websocket.Conn
	send          chan *Message
	hub           Hub
	handler       ConnectionHandler
	self          Connection // The actual connection instance to pass to handlers
	mutex         sync.RWMutex
	writeMutex    sync.Mutex // Protects WebSocket write operations
	isActive      bool
	lastHeartbeat time.Time
	logger        logger.Logger
	config        *ConnectionConfig
}

// NewBaseConnection creates a new base WebSocket connection.
func NewBaseConnection(
	userID uuid.UUID,
	conn *websocket.Conn,
	hub Hub,
	handler ConnectionHandler,
	config *ConnectionConfig,
	logger logger.Logger,
) *BaseConnection {
	if config == nil {
		panic("websocket config is nil")
	}

	base := &BaseConnection{
		id:            uuid.New(),
		userID:        userID,
		conn:          conn,
		send:          make(chan *Message, config.SendBufferSize),
		hub:           hub,
		handler:       handler,
		isActive:      true,
		lastHeartbeat: time.Now(),
		logger:        logger,
		config:        config,
	}

	// By default, self points to the base connection
	base.self = base

	return base
}

// SetSelf sets the actual connection instance that should be passed to handlers.
// This should be called by connection wrappers (like ChatConnection) to ensure
// handlers receive the correct connection type.
func (c *BaseConnection) SetSelf(self Connection) {
	c.self = self
}

// ID returns the connection ID.
func (c *BaseConnection) ID() uuid.UUID {
	return c.id
}

// UserID returns the user ID.
func (c *BaseConnection) UserID() uuid.UUID {
	return c.userID
}

// IsActive returns whether the connection is active.
func (c *BaseConnection) IsActive() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.isActive
}

// Send sends a message to the connection.
func (c *BaseConnection) Send(message *Message) error {
	if !c.IsActive() {
		return ErrConnectionClosed
	}

	select {
	case c.send <- message:
		return nil
	default:
		c.logger.Warn("Send buffer full, closing connection", "connection_id", c.id)

		err := c.Close()
		if err != nil {
			return err
		}

		return ErrSendBufferFull
	}
}

// Close closes the connection.
func (c *BaseConnection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isActive {
		return nil
	}

	c.isActive = false

	// Close the send channel to signal writePump to exit
	// writePump will handle sending the CloseMessage
	close(c.send)

	// Close the underlying network connection
	// writePump will exit when it tries to write after this
	err := c.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

// UpdateHeartbeat updates the last heartbeat timestamp.
func (c *BaseConnection) UpdateHeartbeat() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastHeartbeat = time.Now()
}

// GetLastHeartbeat returns the last heartbeat timestamp.
func (c *BaseConnection) GetLastHeartbeat() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.lastHeartbeat
}

// Start starts the connection's read and write pumps.
func (c *BaseConnection) Start(ctx context.Context) {
	// Notify handler of connection
	if err := c.handler.OnConnect(c.self); err != nil {
		c.logger.Error("Connection handler failed", "error", err)

		err = c.Close()
		if err != nil {
			return
		}

		return
	}

	// Start read and write pumps
	go c.writePump(ctx)
	go c.readPump(ctx)
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *BaseConnection) readPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c.self)

		err := c.Close()
		if err != nil {
			return
		}

		c.handler.OnDisconnect(c.self)
	}()

	c.conn.SetReadLimit(c.config.MaxMessageSize)

	// Give a grace period for initial connection before starting ping/pong cycle
	// This helps with WebSocket clients that don't immediately handle ping/pong
	gracePeriod := c.config.GracePeriod
	initialDeadline := time.Now().Add(gracePeriod)
	c.logger.Debug("Setting initial read deadline with grace period",
		"connection_id", c.id,
		"deadline", initialDeadline,
		"grace_period_seconds", gracePeriod.Seconds(),
		"normal_pong_wait_seconds", c.config.PongWait.Seconds())

	err := c.conn.SetReadDeadline(initialDeadline)
	if err != nil {
		c.logger.Error("Failed to set read deadline", "error", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		c.logger.Debug("Pong message received", "connection_id", c.id)
		c.UpdateHeartbeat()

		newDeadline := time.Now().Add(c.config.PongWait)
		c.logger.Debug("Extending read deadline after pong",
			"connection_id", c.id,
			"new_deadline", newDeadline)

		err = c.conn.SetReadDeadline(newDeadline)
		if err != nil {
			c.logger.Error("Failed to extend read deadline", "error", err)
			return err
		}

		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var message Message

			c.logger.Debug("Waiting for WebSocket message", "connection_id", c.id)

			if err = c.conn.ReadJSON(&message); err != nil {
				c.logger.Debug("WebSocket read error occurred",
					"connection_id", c.id,
					"error", err,
					"error_type", fmt.Sprintf("%T", err))

				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
				) {
					c.logger.Error("WebSocket error", "error", err)
					c.handler.OnError(c.self, err)
				} else {
					c.logger.Debug("WebSocket connection closed normally", "error", err)
				}

				return
			}

			c.logger.Debug("WebSocket message received",
				"connection_id", c.id,
				"message_type", message.Type,
				"message_id", message.ID)

			c.UpdateHeartbeat()

			// Pass the actual connection instance (c.self) to handlers, not the base connection
			if err = c.handler.OnMessage(c.self, &message); err != nil {
				c.logger.Error("Message handler error", "error", err)
				c.handler.OnError(c.self, err)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *BaseConnection) writePump(ctx context.Context) {
	if c.config == nil {
		c.logger.Error("Configuration is nil in writePump - connection cannot function")
		return
	}

	ticker := time.NewTicker(c.config.PingPeriod)

	defer func() {
		ticker.Stop()

		err := c.Close()
		if err != nil {
			c.logger.Error("Close connection error", "error", err)
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.send:
			// Check if connection is still active before writing
			c.mutex.RLock()
			isActive := c.isActive
			c.mutex.RUnlock()

			if !isActive {
				return
			}

			c.writeMutex.Lock()

			// Set write deadline
			err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err != nil {
				c.writeMutex.Unlock()
				return
			}

			if !ok {
				// Send channel was closed, send close message and exit
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.writeMutex.Unlock()

				if err != nil {
					c.logger.Debug(
						"Write close message error (connection likely already closed)",
						"error",
						err,
					)
				}

				return
			}

			if err = c.conn.WriteJSON(message); err != nil {
				c.writeMutex.Unlock()
				c.logger.Error("Write message error", "error", err)

				return
			}

			c.writeMutex.Unlock()

		case <-ticker.C:
			// Check if connection is still active before sending ping
			c.mutex.RLock()
			isActive := c.isActive
			c.mutex.RUnlock()

			if !isActive {
				c.logger.Debug("Skipping ping - connection not active", "connection_id", c.id)
				return
			}

			c.logger.Debug("Sending ping message", "connection_id", c.id)

			c.writeMutex.Lock()

			err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err != nil {
				c.logger.Error("Failed to set write deadline for ping", "error", err)
				c.writeMutex.Unlock()

				return
			}

			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.writeMutex.Unlock()
				c.logger.Debug("Write ping message error (connection likely closed)", "error", err)

				return
			}

			c.logger.Debug("Ping message sent successfully", "connection_id", c.id)
			c.writeMutex.Unlock()
		}
	}
}
