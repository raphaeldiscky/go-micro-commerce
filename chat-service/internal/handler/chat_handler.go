package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	chatDto "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// ChatHandler handles chat-related HTTP requests.
type ChatHandler struct {
	chatService service.ChatService
	hub         *websocket.ChatHub
	logger      logger.Logger
}

// NewChatHandler creates a new ChatHandler instance.
func NewChatHandler(
	chatService service.ChatService,
	hub *websocket.ChatHub,
	appLogger logger.Logger,
) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		hub:         hub,
		logger:      appLogger,
	}
}

// CreateConversation creates a new conversation.
func (h *ChatHandler) CreateConversation(c echo.Context) error {
	var req chatDto.CreateConversationRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("Failed to bind request", "error", err)
		return err
	}

	if err := c.Validate(&req); err != nil {
		h.logger.Info("Failed to validate request", "error", err)
		return err
	}

	userID, userType := h.getUserInfo(c)

	result, err := h.chatService.CreateConversation(
		c.Request().Context(),
		userID,
		userType,
		&req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, result)
}

// GetConversation retrieves a conversation by ID.
func (h *ChatHandler) GetConversation(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	result, err := h.chatService.GetConversation(c.Request().Context(), conversationID, userID)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
}

// GetUserConversations retrieves all conversations for a user.
func (h *ChatHandler) GetUserConversations(c echo.Context) error {
	userID, userType := h.getUserInfo(c)

	result, err := h.chatService.GetUserConversations(c.Request().Context(), userID, userType)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
}

// GetMessages retrieves messages from a conversation with cursor-based pagination.
func (h *ChatHandler) GetMessages(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	limit := int(pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		pkgconstant.DefaultMinLimit,
		pkgconstant.DefaultMaxLimit,
	))

	afterCursor := c.QueryParam("after")
	beforeCursor := c.QueryParam("before")

	messages, paging, err := h.chatService.GetConversationMessagesWithCursor(
		c.Request().Context(),
		conversationID,
		userID,
		limit,
		afterCursor,
		beforeCursor,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOKCursorPagination(c, messages, paging)
}

// JoinConversation adds a participant to a conversation.
func (h *ChatHandler) JoinConversation(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	userID, userType := h.getUserInfo(c)

	role := constant.ParticipantRoleParticipant
	if userType == constant.UserTypeAdmin {
		role = constant.ParticipantRoleModerator
	}

	result, err := h.chatService.JoinConversation(
		c.Request().Context(),
		conversationID,
		userID,
		userType,
		role,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
}

// GetParticipants retrieves participants in a conversation.
func (h *ChatHandler) GetParticipants(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	result, err := h.chatService.GetConversationParticipants(
		c.Request().Context(),
		conversationID,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
}

// GetOnlineUsers retrieves a list of currently online users.
func (h *ChatHandler) GetOnlineUsers(c echo.Context) error {
	userID, _ := h.getUserInfo(c)

	// Get all active connections from the hub
	onlineUsers := h.hub.GetOnlineUsers()

	// Filter out the current user if needed
	filteredUsers := make([]uuid.UUID, 0, len(onlineUsers))
	for _, id := range onlineUsers {
		if id != userID {
			filteredUsers = append(filteredUsers, id)
		}
	}

	response := chatDto.OnlineUsersResponse{
		OnlineUsers: filteredUsers,
		Count:       len(filteredUsers),
	}

	return echoutils.ResponseOK(c, response)
}

// getUserInfo extracts user information from Echo context.
func (h *ChatHandler) getUserInfo(c echo.Context) (uuid.UUID, constant.UserType) {
	userID := echoutils.GetUserIDFromContext(c)
	roles := echoutils.GetRolesFromContext(c)

	// Determine user type based on roles
	userType := constant.UserTypeUser

	if len(roles) > 0 {
		for _, role := range roles {
			if role == pkgconstant.RoleAdmin {
				userType = constant.UserTypeAdmin
				break
			}
		}
	}

	return userID, userType
}
