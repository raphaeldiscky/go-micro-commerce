// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"net/http"
)

// TransportType defines the proxy transport type.
type TransportType int

const (
	// TransportHTTP is for standard HTTP proxying.
	TransportHTTP TransportType = iota
	// TransportWebSocket is for WebSocket proxying.
	TransportWebSocket
	// TransportSSE is for Server-Sent Events proxying.
	TransportSSE
	// TransportConnectRPC is for gRPC/ConnectRPC proxying.
	TransportConnectRPC
)

// ProxyConfig holds configuration for creating a proxy handler.
type ProxyConfig struct {
	ServiceName string
	Path        string // empty = derive from request
	Transport   TransportType
}

// ProxyResponse represents a proxy response for HTTP-based transports.
type ProxyResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
}
