package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// ConnectionResponse represents the response for connection operations.
type ConnectionResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ConnectionID  string    `json:"connection_id"`
	SocketID      string    `json:"socket_id"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	IsActive      bool      `json:"is_active"`
}

// ConnectionRequest represents the request to establish a WebSocket connection.
type ConnectionRequest struct {
	UserAgent string `json:"user_agent,omitempty"`
	ClientIP  string `json:"client_ip,omitempty"`
}

// ChatConnectionResponse represents the response for chat connection establishment.
type ChatConnectionResponse struct {
	NodeAddress string            `json:"node_address"`
	UserID      uuid.UUID         `json:"user_id"`
	UserType    constant.UserType `json:"user_type"`
}

// AuthTokenClaims represents the JWT claims from auth service tokens.
type AuthTokenClaims struct {
	jwt.RegisteredClaims

	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Roles    []string  `json:"roles"`
	IsActive bool      `json:"is_active"`
}

// NodeHealthResponse represents the health status of a chat node.
type NodeHealthResponse struct {
	NodeID         string    `json:"node_id"`
	Address        string    `json:"address"`
	Status         string    `json:"status"`
	Connections    int       `json:"connections"`
	MaxConnections int       `json:"max_connections"`
	LastSeen       time.Time `json:"last_seen"`
}
