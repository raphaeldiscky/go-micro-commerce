package handler

import (
	"time"

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

// SendMessage sends a message in a conversation.
func (h *ChatHandler) SendMessage(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.CreateMessageRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	result, err := h.chatService.SendMessage(
		c.Request().Context(),
		&req,
		userID,
		conversationID,
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

	page := int(pageutils.ParseQueryInt64(
		c,
		"page",
		pkgconstant.DefaultPage,
		pkgconstant.DefaultMinPage,
		pkgconstant.DefaultMaxPage,
	))

	// Convert page to offset
	offset := (page - 1) * limit

	messages, paging, err := h.chatService.GetConversationMessages(
		c.Request().Context(),
		conversationID,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return err
	}

	paging.Links = pageutils.NewLinks(
		c.Request(),
		paging.Page,
		paging.Size,
		paging.TotalPage,
	)

	return echoutils.ResponseOKPagination(c, messages, paging)
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

// UpdatePresence updates a user's presence status.
func (h *ChatHandler) UpdatePresence(c echo.Context) error {
	var req chatDto.UpdatePresenceRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	// Create and broadcast presence message
	presenceMsg, err := websocket.NewPresenceMessage(
		userID,
		req.Status,
		constant.WebSocketEventTypeConnect,
	)
	if err != nil {
		return err
	}

	// Broadcast to all connections
	err = h.hub.Broadcast(presenceMsg, nil)
	if err != nil {
		return err
	}

	response := chatDto.PresenceUpdateResponse{
		UserID:  userID,
		Status:  req.Status,
		Message: "Presence updated successfully",
	}

	return echoutils.ResponseOK(c, response)
}

// SendTypingIndicator sends a typing indicator for a conversation.
func (h *ChatHandler) SendTypingIndicator(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.TypingIndicatorRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	// Create and broadcast typing message
	typingMsg, err := websocket.NewTypingMessage(
		conversationID,
		userID,
		req.IsTyping,
	)
	if err != nil {
		return err
	}

	// Broadcast to conversation participants (excluding sender)
	err = h.hub.BroadcastToConversation(conversationID, typingMsg, userID)
	if err != nil {
		return err
	}

	response := chatDto.TypingIndicatorResponse{
		ConversationID: conversationID,
		UserID:         userID,
		IsTyping:       req.IsTyping,
		Message:        "Typing indicator sent successfully",
	}

	return echoutils.ResponseOK(c, response)
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

// SendDeliveryReceipt sends a delivery receipt for a message.
func (h *ChatHandler) SendDeliveryReceipt(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.DeliveryReceiptRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	// Get current timestamp
	deliveredAt := time.Now()

	// Create and broadcast delivery receipt message
	receiptMsg, err := websocket.NewDeliveryReceiptMessage(
		req.MessageID,
		conversationID,
		userID,
		deliveredAt.Unix(),
	)
	if err != nil {
		return err
	}

	// Broadcast to conversation participants (excluding recipient)
	err = h.hub.BroadcastToConversation(conversationID, receiptMsg, userID)
	if err != nil {
		return err
	}

	response := chatDto.DeliveryReceiptResponse{
		MessageID:      req.MessageID,
		ConversationID: conversationID,
		RecipientID:    userID,
		DeliveredAt:    deliveredAt,
		Message:        "Delivery receipt sent successfully",
	}

	return echoutils.ResponseOK(c, response)
}

// SendReadReceipt sends a read receipt for a message.
func (h *ChatHandler) SendReadReceipt(c echo.Context) error {
	conversationIDStr := c.Param("conversationID")

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return err
	}

	var req chatDto.ReadReceiptRequest
	if err = c.Bind(&req); err != nil {
		return err
	}

	if err = c.Validate(&req); err != nil {
		return err
	}

	userID, _ := h.getUserInfo(c)

	// Get current timestamp
	readAt := time.Now()

	// Create and broadcast read receipt message
	receiptMsg, err := websocket.NewReadReceiptMessage(
		req.MessageID,
		conversationID,
		userID,
		readAt.Unix(),
	)
	if err != nil {
		return err
	}

	// Broadcast to conversation participants (excluding reader)
	err = h.hub.BroadcastToConversation(conversationID, receiptMsg, userID)
	if err != nil {
		return err
	}

	response := chatDto.ReadReceiptResponse{
		MessageID:      req.MessageID,
		ConversationID: conversationID,
		ReaderID:       userID,
		ReadAt:         readAt,
		Message:        "Read receipt sent successfully",
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
