// Package handler provides HTTP handlers for the chat service.
package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
)

// ConnectionHandler handles WebSocket connection management requests.
type ConnectionHandler struct {
	connectionService service.ConnectionService
	logger            logger.Logger
}

// NewConnectionHandler creates a new connection handler.
func NewConnectionHandler(
	connectionService service.ConnectionService,
	logger logger.Logger,
) *ConnectionHandler {
	return &ConnectionHandler{
		connectionService: connectionService,
		logger:            logger,
	}
}

// RequestConnection handles requests for WebSocket connection establishment.
// POST /v1/connect.
func (h *ConnectionHandler) RequestConnection(c echo.Context) error {
	userID := echoutils.GetUserIDFromContext(c)
	roles := echoutils.GetRolesFromContext(c)

	if len(roles) == 0 {
		h.logger.Error("No roles found in context", "user_id", userID)
		return httperror.NewUnauthorizedError("No roles found in context")
	}

	userType := constant.UserType(roles[0]) // Convert to proper type

	// Parse request body (optional for additional metadata)
	var req dto.ConnectionRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Invalid request body", "error", err)
		return err
	}

	if err := c.Validate(&req); err != nil {
		h.logger.Error("Invalid request body", "error", err)
		return err
	}

	// Request connection from service
	response, err := h.connectionService.RequestConnection(c.Request().Context(), userID, userType)
	if err != nil {
		h.logger.Error("Failed to request connection", "error", err, "user_id", userID)
		return err
	}

	h.logger.Info("Connection requested successfully",
		"user_id", userID,
		"user_type", userType,
		"node_address", response.NodeAddress)

	return echoutils.ResponseOK(c, response)
}

// GetNodeHealth returns health status of available chat nodes.
// GET /v1/nodes/health.
func (h *ConnectionHandler) GetNodeHealth(c echo.Context) error {
	nodes, err := h.connectionService.GetNodeHealth(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to get node health", "error", err)
		return err
	}

	return echoutils.ResponseOK(c, map[string]interface{}{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// ValidateTicket validates a connection ticket (for testing purposes).
// POST /v1/validate-ticket.
func (h *ConnectionHandler) ValidateTicket(c echo.Context) error {
	var req struct {
		Ticket string `json:"ticket" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	claims, err := h.connectionService.ValidateConnectionTicket(c.Request().Context(), req.Ticket)
	if err != nil {
		h.logger.Error("Failed to validate ticket", "error", err)
		return err
	}

	return echoutils.ResponseOK(c, claims)
}
