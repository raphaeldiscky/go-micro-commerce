package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
}

// NewChatHandler creates a new ChatHandler instance.
func NewChatHandler(chatService service.ChatService, _ *websocket.ChatHub) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
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

// CreateConversation creates a new conversation.
func (h *ChatHandler) CreateConversation(c echo.Context) error {
	var req chatDto.CreateConversationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
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

// SendMessage sends a message in a conversation.
func (h *ChatHandler) SendMessage(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	req.ConversationID = conversationID

	userID, userType := h.getUserInfo(c)

	result, err := h.chatService.SendMessage(
		c.Request().Context(),
		&req,
		userID,
		userType,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseCreated(c, result)
}

// GetMessages retrieves messages from a conversation with pagination.
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

	offset := int(pageutils.ParseQueryInt64(
		c,
		"offset",
		0,
		0,
		constant.DefaultConversationLimit,
	))

	result, err := h.chatService.GetConversationMessages(
		c.Request().Context(),
		conversationID,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
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

// UpdateConversationStatus updates the status of a conversation.
func (h *ChatHandler) UpdateConversationStatus(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.UpdateConversationStatusRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.chatService.UpdateConversationStatus(
		c.Request().Context(),
		conversationID,
		&req,
	)
	if err != nil {
		return err
	}

	return echoutils.ResponseOK(c, result)
}
