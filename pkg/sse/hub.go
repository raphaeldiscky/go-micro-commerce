package sse

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"
)

// Hub manages SSE connections and broadcasts messages.
type Hub struct {
	connections        map[uuid.UUID]*Connection               // All connections by ID
	userConns          map[uuid.UUID]map[uuid.UUID]*Connection // User connections by user ID
	register           chan *Connection
	unregister         chan *Connection
	broadcast          chan *BroadcastRequest
	mutex              sync.RWMutex
	logger             logger.Logger
	done               chan struct{}
	eventBus           rediseventbus.EventBus
	instanceID         string
	userChannelBuilder func(uuid.UUID) string     // Function to build channel name for a user
	eventHandler       rediseventbus.EventHandler // Handler for cross-instance events
	subscribedUsers    map[uuid.UUID]int          // userID -> connection count (for subscription tracking)
}

// BroadcastRequest represents a request to broadcast a message.
type BroadcastRequest struct {
	UserID  *uuid.UUID // If nil, broadcast to all users
	Message *Message
}

// NewHub creates a new SSE hub.
func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		connections:     make(map[uuid.UUID]*Connection),
		userConns:       make(map[uuid.UUID]map[uuid.UUID]*Connection),
		register:        make(chan *Connection),
		unregister:      make(chan *Connection),
		broadcast:       make(chan *BroadcastRequest, constant.SSEBroadcastBufferSize),
		logger:          logger,
		done:            make(chan struct{}),
		subscribedUsers: make(map[uuid.UUID]int),
	}
}

// Register registers a new connection.
func (h *Hub) Register(conn *Connection) {
	select {
	case h.register <- conn:
	case <-h.done:
		h.logger.Warn("Hub is shutting down, cannot register connection")
	}
}

// Unregister unregisters a connection.
func (h *Hub) Unregister(conn *Connection) {
	select {
	case h.unregister <- conn:
	case <-h.done:
		// Hub is shutting down, connection will be cleaned up
	}
}

// BroadcastToUser broadcasts a message to a specific user.
func (h *Hub) BroadcastToUser(userID uuid.UUID, message *Message) error {
	request := &BroadcastRequest{
		UserID:  &userID,
		Message: message,
	}

	select {
	case h.broadcast <- request:
		return nil
	case <-h.done:
		return ErrHubShutdown
	}
}

// BroadcastToAll broadcasts a message to all connected users.
func (h *Hub) BroadcastToAll(message *Message) error {
	request := &BroadcastRequest{
		UserID:  nil,
		Message: message,
	}

	select {
	case h.broadcast <- request:
		return nil
	case <-h.done:
		return ErrHubShutdown
	}
}

// GetUserConnections retrieves all connections for a user.
func (h *Hub) GetUserConnections(userID uuid.UUID) []*Connection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	userMap, exists := h.userConns[userID]
	if !exists {
		return nil
	}

	connections := make([]*Connection, 0, len(userMap))
	for _, conn := range userMap {
		if conn.IsActive() {
			connections = append(connections, conn)
		}
	}

	return connections
}

// GetConnectionCount returns the total number of active connections.
func (h *Hub) GetConnectionCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.connections)
}

// GetUserCount returns the number of unique users connected.
func (h *Hub) GetUserCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.userConns)
}

// Run starts the hub and processes connection events.
func (h *Hub) Run(ctx context.Context) {
	cleanupTicker := time.NewTicker(constant.SSECleanupTicker)
	defer cleanupTicker.Stop()

	h.logger.Info("SSE Hub started")

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
func (h *Hub) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutting down SSE hub...")

	// Signal shutdown
	close(h.done)

	// Close all connections
	done := make(chan struct{})

	go func() {
		h.closeAllConnections()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Info("All SSE connections closed successfully")
	case <-ctx.Done():
		h.logger.Warn("Shutdown timeout reached, some connections may not have closed gracefully")
	}

	return nil
}

// handleRegister handles connection registration.
func (h *Hub) handleRegister(conn *Connection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	userID := conn.UserID()
	connID := conn.ID()

	// Add to connections map
	h.connections[connID] = conn

	// Add to user connections map
	if h.userConns[userID] == nil {
		h.userConns[userID] = make(map[uuid.UUID]*Connection)
	}

	h.userConns[userID][connID] = conn

	// Subscribe to user's Redis channel if this is their first connection
	if h.eventBus != nil && h.userChannelBuilder != nil {
		if h.subscribedUsers[userID] == 0 {
			if err := h.subscribeToUserChannel(userID); err != nil {
				h.logger.Error("Failed to subscribe to user channel",
					"user_id", userID,
					"error", err)
			}
		}

		h.subscribedUsers[userID]++
	}

	h.logger.Info("SSE connection registered",
		"connection_id", connID,
		"user_id", userID,
		"total_connections", len(h.connections),
		"subscribed_users", len(h.subscribedUsers))
}

// handleUnregister handles connection unregistration.
func (h *Hub) handleUnregister(conn *Connection) {
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

	// Unsubscribe from user's Redis channel if this was their last connection
	h.handleRedisUnsubscription(userID)

	h.logger.Info("SSE connection unregistered",
		"connection_id", connID,
		"user_id", userID,
		"total_connections", len(h.connections),
		"subscribed_users", len(h.subscribedUsers))
}

// handleBroadcast handles message broadcasting.
func (h *Hub) handleBroadcast(request *BroadcastRequest) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var targets []*Connection

	if request.UserID != nil {
		if userMap, exists := h.userConns[*request.UserID]; exists {
			for _, conn := range userMap {
				targets = append(targets, conn)
			}
		} else {
			h.logger.Debug("No connections found for user",
				"user_id", *request.UserID)

			return
		}
	} else {
		for _, conn := range h.connections {
			targets = append(targets, conn)
		}
	}

	sentCount := h.broadcastToTargets(targets, request.Message)

	h.logger.Debug("SSE message broadcasted",
		"message_id", request.Message.ID,
		"event", request.Message.Event,
		"recipients", sentCount)
}

// broadcastToTargets sends a message to multiple connections and returns the count of successful sends.
func (h *Hub) broadcastToTargets(targets []*Connection, message *Message) int {
	sentCount := 0

	for _, conn := range targets {
		if !conn.IsActive() {
			continue
		}

		if err := conn.Send(message); err != nil {
			h.logger.Warn("Failed to send message to connection",
				"connection_id", conn.ID(),
				"user_id", conn.UserID(),
				"error", err)

			continue
		}

		sentCount++
	}

	return sentCount
}

// cleanupStaleConnections removes inactive connections.
func (h *Hub) cleanupStaleConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	staleThreshold := time.Now().Add(-5 * time.Minute)
	staleConnections := make([]*Connection, 0)

	for _, conn := range h.connections {
		if !conn.IsActive() || conn.GetLastHeartbeat().Before(staleThreshold) {
			staleConnections = append(staleConnections, conn)
		}
	}

	for _, conn := range staleConnections {
		h.logger.Info("Cleaning up stale SSE connection", "connection_id", conn.ID())

		if err := conn.Close(); err != nil {
			h.logger.Error("Failed to close connection", "connection_id", conn.ID(), "error", err)
		}
	}
}

// closeAllConnections closes all active connections.
func (h *Hub) closeAllConnections() {
	h.mutex.RLock()

	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}

	h.mutex.RUnlock()

	for _, conn := range connections {
		if err := conn.Close(); err != nil {
			h.logger.Error("Failed to close connection", "connection_id", conn.ID(), "error", err)
		}
	}
}

// shutdown performs internal shutdown tasks.
func (h *Hub) shutdown() {
	h.logger.Info("SSE Hub shutdown initiated")
	h.closeAllConnections()
}

// SetEventBus configures the hub to receive events from other instances via Redis sharded pub/sub.
// Uses per-user channels with Redis 7.0+ native sharded pub/sub (SSUBSCRIBE).
// Channels are subscribed/unsubscribed dynamically as users connect/disconnect.
// The eventHandler will be called for all cross-instance events (should call BroadcastToUser).
func (h *Hub) SetEventBus(
	eventBus rediseventbus.EventBus,
	instanceID string,
	userChannelBuilder func(uuid.UUID) string,
	eventHandler rediseventbus.EventHandler,
) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.eventBus = eventBus
	h.instanceID = instanceID
	h.userChannelBuilder = userChannelBuilder
	h.eventHandler = eventHandler

	h.logger.Info("EventBus configured for SSE Hub",
		"instance_id", instanceID,
		"using_sharded_pubsub", true)
}

// subscribeToUserChannel subscribes to a user's Redis sharded channel.
// Must be called with hub mutex locked.
func (h *Hub) subscribeToUserChannel(userID uuid.UUID) error {
	if h.eventBus == nil || h.userChannelBuilder == nil || h.eventHandler == nil {
		return nil
	}

	channel := h.userChannelBuilder(userID)

	// Create wrapper handler that filters by instance ID
	wrappedHandler := func(ctx context.Context, event rediseventbus.Event) error {
		// Skip events from our own instance to avoid duplicate delivery
		if event.GetSourceInstanceID() == h.instanceID {
			h.logger.Debug("Skipping event from own instance",
				"instance_id", h.instanceID,
				"user_id", userID,
				"event_type", event.GetType())

			return nil
		}

		h.logger.Debug("Received cross-instance event",
			"source_instance_id", event.GetSourceInstanceID(),
			"user_id", userID,
			"event_type", event.GetType())

		// Delegate to service-specific event handler
		return h.eventHandler(ctx, event)
	}

	// Use SSubscribe for Redis 7.0+ sharded pub/sub
	if err := h.eventBus.SSubscribe(channel, wrappedHandler); err != nil {
		return err
	}

	h.logger.Info("Subscribed to user channel",
		"user_id", userID,
		"channel", channel)

	return nil
}

// unsubscribeFromUserChannel unsubscribes from a user's Redis sharded channel.
// Must be called with hub mutex locked.
func (h *Hub) unsubscribeFromUserChannel(userID uuid.UUID) error {
	if h.eventBus == nil || h.userChannelBuilder == nil {
		return nil
	}

	channel := h.userChannelBuilder(userID)

	// Use SUnsubscribe for Redis 7.0+ sharded pub/sub
	if err := h.eventBus.SUnsubscribe(channel); err != nil {
		return err
	}

	h.logger.Info("Unsubscribed from user channel",
		"user_id", userID,
		"channel", channel)

	return nil
}

// handleRedisUnsubscription handles Redis unsubscription when a user connection is removed.
// Must be called with hub mutex locked.
func (h *Hub) handleRedisUnsubscription(userID uuid.UUID) {
	if h.eventBus == nil || h.userChannelBuilder == nil {
		return
	}

	count, exists := h.subscribedUsers[userID]
	if !exists {
		return
	}

	h.subscribedUsers[userID]--

	if h.subscribedUsers[userID] == 0 {
		delete(h.subscribedUsers, userID)

		if err := h.unsubscribeFromUserChannel(userID); err != nil {
			h.logger.Error("Failed to unsubscribe from user channel",
				"user_id", userID,
				"error", err)
		}
	}

	h.logger.Debug("Decremented user subscription count",
		"user_id", userID,
		"previous_count", count,
		"new_count", h.subscribedUsers[userID])
}
