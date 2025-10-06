package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Hub represents a universal WebSocket hub interface.
type Hub interface {
	// Register registers a new connection
	Register(conn Connection)
	// Unregister unregisters a connection
	Unregister(conn Connection)
	// Broadcast broadcasts a message to connections
	Broadcast(message *Message, filter ConnectionFilter) error
	// BroadcastToChannel broadcasts a message to a specific channel
	BroadcastToChannel(channel string, message *Message) error
	// BroadcastToUser broadcasts a message to a specific user
	BroadcastToUser(userID uuid.UUID, message *Message) error
	// GetConnection retrieves a connection by ID
	GetConnection(connID uuid.UUID) (Connection, bool)
	// GetUserConnections retrieves all connections for a user
	GetUserConnections(userID uuid.UUID) []Connection
	// GetChannelConnections retrieves all connections in a channel
	GetChannelConnections(channel string) []Connection
	// Run starts the hub
	Run(ctx context.Context)
	// Shutdown gracefully shuts down the hub
	Shutdown(ctx context.Context) error
}

// ConnectionFilter defines a function type for filtering connections.
type ConnectionFilter func(Connection) bool

// BaseHub provides a base implementation of the Hub interface.
type BaseHub struct {
	connections        map[uuid.UUID]Connection               // All connections by ID
	userConns          map[uuid.UUID]map[uuid.UUID]Connection // User connections by user ID
	channels           map[string]map[uuid.UUID]Connection    // Channel connections by channel name
	channelSubscribers map[string]map[string]chan<- *Message  // External channel subscribers (for GraphQL subscriptions)
	register           chan Connection
	unregister         chan Connection
	broadcast          chan *BroadcastRequest
	mutex              sync.RWMutex
	logger             logger.Logger
	done               chan struct{}
}

// BroadcastRequest represents a request to broadcast a message.
type BroadcastRequest struct {
	Message *Message
	Filter  ConnectionFilter
}

// NewBaseHub creates a new base hub instance.
func NewBaseHub(logger logger.Logger) *BaseHub {
	return &BaseHub{
		connections:        make(map[uuid.UUID]Connection),
		userConns:          make(map[uuid.UUID]map[uuid.UUID]Connection),
		channels:           make(map[string]map[uuid.UUID]Connection),
		channelSubscribers: make(map[string]map[string]chan<- *Message),
		register:           make(chan Connection),
		unregister:         make(chan Connection),
		broadcast:          make(chan *BroadcastRequest),
		logger:             logger,
		done:               make(chan struct{}),
	}
}

// Register registers a new connection.
func (h *BaseHub) Register(conn Connection) {
	select {
	case h.register <- conn:
	case <-h.done:
		h.logger.Warn("Hub is shutting down, cannot register connection")
	}
}

// Unregister unregisters a connection.
func (h *BaseHub) Unregister(conn Connection) {
	select {
	case h.unregister <- conn:
	case <-h.done:
		// Hub is shutting down, connection will be cleaned up
	}
}

// Broadcast broadcasts a message to connections matching the filter.
func (h *BaseHub) Broadcast(message *Message, filter ConnectionFilter) error {
	request := &BroadcastRequest{
		Message: message,
		Filter:  filter,
	}

	select {
	case h.broadcast <- request:
		return nil
	case <-h.done:
		return ErrHubShutdown
	}
}

// BroadcastToChannel broadcasts a message to a specific channel.
func (h *BaseHub) BroadcastToChannel(channel string, message *Message) error {
	filter := func(conn Connection) bool {
		return h.isConnectionInChannel(conn, channel)
	}

	return h.Broadcast(message, filter)
}

// BroadcastToUser broadcasts a message to a specific user.
func (h *BaseHub) BroadcastToUser(userID uuid.UUID, message *Message) error {
	filter := func(conn Connection) bool {
		return conn.UserID() == userID
	}

	return h.Broadcast(message, filter)
}

// SubscribeToChannel registers an external message channel to receive all broadcasts to a specific channel.
// This is primarily used for GraphQL subscriptions to bridge WebSocket events.
// Returns an unsubscribe function to remove the subscription.
func (h *BaseHub) SubscribeToChannel(channelName string, messageChan chan<- *Message) func() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	subID := uuid.New().String()

	if h.channelSubscribers[channelName] == nil {
		h.channelSubscribers[channelName] = make(map[string]chan<- *Message)
	}

	h.channelSubscribers[channelName][subID] = messageChan

	h.logger.Info("Channel subscriber added",
		"channel", channelName,
		"subscriber_id", subID)

	// Return unsubscribe function
	return func() {
		h.mutex.Lock()
		defer h.mutex.Unlock()

		if subs, exists := h.channelSubscribers[channelName]; exists {
			delete(subs, subID)

			if len(subs) == 0 {
				delete(h.channelSubscribers, channelName)
			}
		}

		h.logger.Info("Channel subscriber removed",
			"channel", channelName,
			"subscriber_id", subID)
	}
}

// GetConnection retrieves a connection by ID.
func (h *BaseHub) GetConnection(connID uuid.UUID) (Connection, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	conn, exists := h.connections[connID]

	return conn, exists
}

// GetUserConnections retrieves all connections for a user.
func (h *BaseHub) GetUserConnections(userID uuid.UUID) []Connection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	userMap, exists := h.userConns[userID]
	if !exists {
		return nil
	}

	connections := make([]Connection, 0, len(userMap))
	for _, conn := range userMap {
		if conn.IsActive() {
			connections = append(connections, conn)
		}
	}

	return connections
}

// GetChannelConnections retrieves all connections in a channel.
func (h *BaseHub) GetChannelConnections(channel string) []Connection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	channelMap, exists := h.channels[channel]
	if !exists {
		return nil
	}

	connections := make([]Connection, 0, len(channelMap))
	for _, conn := range channelMap {
		if conn.IsActive() {
			connections = append(connections, conn)
		}
	}

	return connections
}

// Run starts the hub and processes connection events.
func (h *BaseHub) Run(ctx context.Context) {
	cleanupTicker := time.NewTicker(constant.WsCleanupTicker)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case conn := <-h.register:
			h.handleRegister(conn)

		case conn := <-h.unregister:
			h.handleUnregister(conn)

		case request := <-h.broadcast:
			h.handleBroadcast(request)

		case <-cleanupTicker.C:
			h.cleanupStaleConnections()
		}
	}
}

// Shutdown gracefully shuts down the hub.
func (h *BaseHub) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutting down WebSocket hub...")

	// Signal shutdown
	close(h.done)

	// Close all connections
	shutdownCtx, cancel := context.WithTimeout(ctx, constant.WsShutdownTimeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		h.closeAllConnections()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Info("All connections closed successfully")
	case <-shutdownCtx.Done():
		h.logger.Warn("Shutdown timeout reached, some connections may not have closed gracefully")
	}

	return nil
}

// handleRegister handles connection registration.
func (h *BaseHub) handleRegister(conn Connection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	userID := conn.UserID()
	connID := conn.ID()

	// Add to connections map
	h.connections[connID] = conn

	// Add to user connections map
	if h.userConns[userID] == nil {
		h.userConns[userID] = make(map[uuid.UUID]Connection)
	}

	h.userConns[userID][connID] = conn

	h.logger.Info("Connection registered",
		"connection_id", connID,
		"user_id", userID,
		"total_connections", len(h.connections))
}

// handleUnregister handles connection unregistration.
func (h *BaseHub) handleUnregister(conn Connection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	userID := conn.UserID()
	connID := conn.ID()

	// Remove from connections map
	delete(h.connections, connID)

	// Remove from user connections map
	if userMap, exists := h.userConns[userID]; exists {
		delete(userMap, connID)

		if len(userMap) == 0 {
			delete(h.userConns, userID)
		}
	}

	// Remove from all channels
	for channelName, channelMap := range h.channels {
		delete(channelMap, connID)

		if len(channelMap) == 0 {
			delete(h.channels, channelName)
		}
	}

	h.logger.Info("Connection unregistered",
		"connection_id", connID,
		"user_id", userID,
		"total_connections", len(h.connections))
}

// handleBroadcast handles message broadcasting.
func (h *BaseHub) handleBroadcast(request *BroadcastRequest) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	sentCount := 0

	for _, conn := range h.connections {
		if !conn.IsActive() {
			continue
		}

		// Apply filter if provided
		if request.Filter != nil && !request.Filter(conn) {
			continue
		}

		if err := conn.Send(request.Message); err != nil {
			h.logger.Warn("Failed to send message to connection",
				"connection_id", conn.ID(),
				"error", err)
		} else {
			sentCount++
		}
	}

	// Forward to channel subscribers if message has a channel
	if request.Message.Channel != nil {
		channelName := *request.Message.Channel
		if subscribers, exists := h.channelSubscribers[channelName]; exists {
			for subID, subChan := range subscribers {
				select {
				case subChan <- request.Message:
					sentCount++
				default:
					h.logger.Warn("Channel subscriber full, dropping message",
						"channel", channelName,
						"subscriber_id", subID)
				}
			}
		}
	}

	h.logger.Debug("Message broadcasted",
		"message_id", request.Message.ID,
		"recipients", sentCount)
}

// JoinChannel adds a connection to a channel.
func (h *BaseHub) JoinChannel(conn Connection, channel string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	connID := conn.ID()

	if h.channels[channel] == nil {
		h.channels[channel] = make(map[uuid.UUID]Connection)
	}

	h.channels[channel][connID] = conn

	h.logger.Info("Connection joined channel",
		"connection_id", connID,
		"channel", channel)
}

// LeaveChannel removes a connection from a channel.
func (h *BaseHub) LeaveChannel(conn Connection, channel string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	connID := conn.ID()

	if channelMap, exists := h.channels[channel]; exists {
		delete(channelMap, connID)

		if len(channelMap) == 0 {
			delete(h.channels, channel)
		}
	}

	h.logger.Info("Connection left channel",
		"connection_id", connID,
		"channel", channel)
}

// isConnectionInChannel checks if a connection is in a specific channel.
func (h *BaseHub) isConnectionInChannel(conn Connection, channel string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	channelMap, exists := h.channels[channel]
	if !exists {
		return false
	}

	_, inChannel := channelMap[conn.ID()]

	return inChannel
}

// cleanupStaleConnections removes inactive connections.
func (h *BaseHub) cleanupStaleConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	staleThreshold := time.Now().Add(-2 * time.Minute)
	staleConnections := make([]Connection, 0)

	for _, conn := range h.connections {
		if !conn.IsActive() || conn.GetLastHeartbeat().Before(staleThreshold) {
			staleConnections = append(staleConnections, conn)
		}
	}

	for _, conn := range staleConnections {
		h.logger.Info("Cleaning up stale connection", "connection_id", conn.ID())

		err := conn.Close()
		if err != nil {
			h.logger.Error("Failed to close connection", "connection_id", conn.ID(), "error", err)
		}
	}
}

// closeAllConnections closes all active connections.
func (h *BaseHub) closeAllConnections() {
	h.mutex.RLock()

	connections := make([]Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}

	h.mutex.RUnlock()

	for _, conn := range connections {
		err := conn.Close()
		if err != nil {
			h.logger.Error("Failed to close connection", "connection_id", conn.ID(), "error", err)
		}
	}
}

// GetConnectionCount returns the total number of active connections.
func (h *BaseHub) GetConnectionCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.connections)
}

// GetUserCount returns the number of unique users connected.
func (h *BaseHub) GetUserCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.userConns)
}

// shutdown performs internal shutdown tasks.
func (h *BaseHub) shutdown() {
	h.logger.Info("Hub shutdown initiated")
	h.closeAllConnections()
}
