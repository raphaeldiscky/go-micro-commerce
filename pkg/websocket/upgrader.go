package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// UpgraderConfig holds configuration for WebSocket upgrader.
type UpgraderConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
	Subprotocols    []string
}

// DefaultUpgraderConfig returns default upgrader configuration.
func DefaultUpgraderConfig() *UpgraderConfig {
	return &UpgraderConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins in default config - should be customized for production
			return true
		},
		Subprotocols: []string{},
	}
}

// NewUpgrader creates a new WebSocket upgrader with the given configuration.
func NewUpgrader(config *UpgraderConfig) *websocket.Upgrader {
	if config == nil {
		config = DefaultUpgraderConfig()
	}

	return &websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
		Subprotocols:    config.Subprotocols,
	}
}

// Upgrade upgrades an HTTP connection to a WebSocket connection.
func Upgrade(
	w http.ResponseWriter,
	r *http.Request,
	config *UpgraderConfig,
) (*websocket.Conn, error) {
	upgrader := NewUpgrader(config)
	return upgrader.Upgrade(w, r, nil)
}
