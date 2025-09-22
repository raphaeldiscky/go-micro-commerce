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

// NewUpgrader creates a new WebSocket upgrader with the given configuration.
func NewUpgrader(config *UpgraderConfig) *websocket.Upgrader {
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
