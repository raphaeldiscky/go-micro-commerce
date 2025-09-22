package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
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

// BaseConnection provides a base implementation of the Connection interface.
type BaseConnection struct {
	id            uuid.UUID
	userID        uuid.UUID
	conn          *websocket.Conn
	send          chan *Message
	hub           Hub
	handler       ConnectionHandler
	mutex         sync.RWMutex
	isActive      bool
	lastHeartbeat time.Time
	logger        logger.Logger
	config        *config.WebsocketServerConfig
}

// NewBaseConnection creates a new base WebSocket connection.
func NewBaseConnection(
	userID uuid.UUID,
	conn *websocket.Conn,
	hub Hub,
	handler ConnectionHandler,
	config *config.WebsocketServerConfig,
	logger logger.Logger,
) *BaseConnection {
	return &BaseConnection{
		id:            uuid.New(),
		userID:        userID,
		conn:          conn,
		send:          make(chan *Message, config.SendBufferSize),
		hub:           hub,
		handler:       handler,
		isActive:      true,
		lastHeartbeat: time.Now(),
		logger:        logger,
	}
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
	close(c.send)

	// Set close handler
	err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
	if err != nil {
		return err
	}

	err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		return err
	}

	err = c.conn.Close()
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
	if err := c.handler.OnConnect(c); err != nil {
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
		c.hub.Unregister(c)

		err := c.Close()
		if err != nil {
			return
		}

		c.handler.OnDisconnect(c)
	}()

	c.conn.SetReadLimit(c.config.MaxMessageSize)

	err := c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	if err != nil {
		return
	}

	c.conn.SetPongHandler(func(string) error {
		c.UpdateHeartbeat()

		err = c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
		if err != nil {
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
			if err = c.conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
				) {
					c.logger.Error("WebSocket error", "error", err)
					c.handler.OnError(c, err)
				}

				return
			}

			c.UpdateHeartbeat()

			if err = c.handler.OnMessage(c, &message); err != nil {
				c.logger.Error("Message handler error", "error", err)
				c.handler.OnError(c, err)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *BaseConnection) writePump(ctx context.Context) {
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
			err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err != nil {
				return
			}

			if !ok {
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					c.logger.Error("Write message error", "error", err)
					return
				}

				return
			}

			if err = c.conn.WriteJSON(message); err != nil {
				c.logger.Error("Write message error", "error", err)
				return
			}

		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait))
			if err != nil {
				return
			}

			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
