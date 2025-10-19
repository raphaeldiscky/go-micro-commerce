// Package service provides business logic services for the notification service.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/subscription"
)

// NotificationRequestEvent is the envelope for notification request events.
type NotificationRequestEvent struct {
	Metadata kafkaevent.Metadata                   `json:"metadata"`
	Payload  kafkaevent.NotificationRequestPayload `json:"payload"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata kafkaevent.Metadata            `json:"metadata"`
	Payload  kafkaevent.UserVerifiedPayload `json:"payload"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata kafkaevent.Metadata                          `json:"metadata"`
	Payload  kafkaevent.EmailVerificationRequestedPayload `json:"payload"`
}

// NotificationEventService handles all notification business logic.
type NotificationEventService interface {
	ProcessNotificationRequest(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessEmailVerificationRequest(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessEmailUserVerified(ctx context.Context, inboxEvent *entity.InboxEvent) error

	// CreateAndBroadcastNotification creates a notification and broadcasts it via SSE/Redis
	CreateAndBroadcastNotification(
		ctx context.Context,
		userID *uuid.UUID, // nil for system notifications to broadcast for all users
		notificationType constant.PushNotificationType,
		title string,
		message string,
		metadata map[string]any,
	) (*dto.NotificationResponse, error)
}

// notificationEventService implements notification business logic.
type notificationEventService struct {
	emailService     EmailService
	notificationRepo repository.NotificationRepository
	sseHub           *sse.Hub
	eventBus         rediseventbus.EventBus
	instanceID       string
	logger           logger.Logger
}

// NewNotificationEventService creates a new notification service instance.
func NewNotificationEventService(
	emailService EmailService,
	notificationRepo repository.NotificationRepository,
	sseHub *sse.Hub,
	eventBus rediseventbus.EventBus,
	instanceID string,
	appLogger logger.Logger,
) NotificationEventService {
	return &notificationEventService{
		emailService:     emailService,
		notificationRepo: notificationRepo,
		sseHub:           sseHub,
		eventBus:         eventBus,
		instanceID:       instanceID,
		logger:           appLogger,
	}
}

// ProcessNotificationRequest handles notification request events.
func (s *notificationEventService) ProcessNotificationRequest(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing notification request event: %s", inboxEvent.MessageID)

	var notificationEvent NotificationRequestEvent
	if err := json.Unmarshal(inboxEvent.Payload, &notificationEvent); err != nil {
		return fmt.Errorf("failed to unmarshal notification request event: %w", err)
	}

	switch notificationEvent.Payload.NotificationType {
	case kafkaevent.NotificationTypeEmail:
		return s.sendEmailNotification(ctx, &notificationEvent.Payload)
	case kafkaevent.NotificationTypeSMS:
		s.logger.Info("SMS notifications not yet implemented")

		return nil
	case kafkaevent.NotificationTypePush:
		return s.sendPushNotification(ctx, &notificationEvent.Payload)
	default:
		return fmt.Errorf(
			"unsupported notification type: %s",
			notificationEvent.Payload.NotificationType,
		)
	}
}

// processEmailEvent is a generic helper for processing email events.
func (s *notificationEventService) processEmailEvent(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
	eventType string,
	unmarshalFn func([]byte) (string, string, error),
) error {
	s.logger.Infof("Processing %s event: %s", eventType, inboxEvent.MessageID)

	email, body, err := unmarshalFn(inboxEvent.Payload)
	if err != nil {
		return err
	}

	subject := getSubjectForEventType(eventType)

	if err = s.emailService.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("failed to send %s email: %w", eventType, err)
	}

	s.logger.Infof("Successfully sent %s email to: %s", eventType, email)

	return nil
}

// getSubjectForEventType returns the appropriate subject for the event type.
func getSubjectForEventType(eventType string) string {
	switch eventType {
	case kafka.EmailVerificationRequestedEventType:
		return constant.SendVerificationEmailSubject
	case kafka.UserVerifiedEventType:
		return constant.UserVerifiedEmailSubject
	default:
		return ""
	}
}

// ProcessEmailVerificationRequest handles email verification request events.
func (s *notificationEventService) ProcessEmailVerificationRequest(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	unmarshalFn := func(payload []byte) (string, string, error) {
		var emailVerificationEvent EmailVerificationRequestedEvent
		if err := json.Unmarshal(payload, &emailVerificationEvent); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal email verification event: %w", err)
		}

		body, err := s.generateVerificationEmail(&emailVerificationEvent.Payload)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate verification email: %w", err)
		}

		return emailVerificationEvent.Payload.Email, body, nil
	}

	return s.processEmailEvent(
		ctx,
		inboxEvent,
		kafka.EmailVerificationRequestedEventType,
		unmarshalFn,
	)
}

// ProcessEmailUserVerified handles user verified events.
func (s *notificationEventService) ProcessEmailUserVerified(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	unmarshalFn := func(payload []byte) (string, string, error) {
		var userVerifiedEvent UserVerifiedEvent
		if err := json.Unmarshal(payload, &userVerifiedEvent); err != nil {
			return "", "", fmt.Errorf("failed to unmarshal user verified event: %w", err)
		}

		body, err := s.generateWelcomeEmail(&userVerifiedEvent.Payload)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate welcome email: %w", err)
		}

		return userVerifiedEvent.Payload.Email, body, nil
	}

	return s.processEmailEvent(ctx, inboxEvent, kafka.UserVerifiedEventType, unmarshalFn)
}

// generateVerificationEmail creates an email verification email body.
func (s *notificationEventService) generateVerificationEmail(
	payload *kafkaevent.EmailVerificationRequestedPayload,
) (string, error) {
	verificationURL := fmt.Sprintf("http://localhost:8080/auth/v1/verify?token=%s", payload.Token)

	templateData := &dto.EmailVerificationTemplateData{
		RecipientName:   payload.Email,
		VerificationURL: verificationURL,
		TokenExpiresAt:  payload.TokenExpiresAt.Format(time.Kitchen),
	}

	return s.emailService.RenderTemplate(constant.TemplateFileEmailVerification, templateData)
}

// generateWelcomeEmail creates a welcome email body.
func (s *notificationEventService) generateWelcomeEmail(
	payload *kafkaevent.UserVerifiedPayload,
) (string, error) {
	templateData := &dto.UserVerifiedTemplateData{
		RecipientName: payload.Email,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileUserVerified, templateData)
}

// sendEmailNotification sends an email notification.
func (s *notificationEventService) sendEmailNotification(
	ctx context.Context,
	payload *kafkaevent.NotificationRequestPayload,
) error {
	s.logger.Infof("Sending email to %s with subject: %s", payload.RecipientEmail, payload.Subject)

	emailBody, err := s.generateEmailBody(payload)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	if err = s.emailService.SendEmail(ctx, payload.RecipientEmail, payload.Subject, emailBody); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", payload.RecipientEmail, err)
	}

	s.logger.Infof("Successfully sent email to %s", payload.RecipientEmail)
	s.logger.Printf("EMAIL SENT: To=%s, Subject=%s, Template=%s",
		payload.RecipientEmail, payload.Subject, payload.TemplateID)

	return nil
}

// generateEmailBody generates the email body based on template ID and data.
func (s *notificationEventService) generateEmailBody(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	s.logger.Infof(
		"Processing template ID: '%s' for email: %s",
		payload.TemplateID,
		payload.RecipientEmail,
	)

	switch payload.TemplateID {
	case pkgconstant.TemplateOrderConfirmed:
		return s.generateOrderConfirmedEmail(payload)
	case pkgconstant.TemplateOrderShipped:
		return s.generateOrderShippedEmail(payload)
	case pkgconstant.TemplateOrderCanceled:
		return s.generateOrderCancelledEmail(payload)
	case pkgconstant.TemplateOrderPaymentExpired:
		return s.generateOrderPaymentExpiredEmail(payload)
	case pkgconstant.TemplateOrderDelivered:
		return s.generateOrderDeliveredEmail(payload)
	case pkgconstant.TemplateOrderPaymentRequired:
		return s.generateOrderPaymentRequiredEmail(payload)
	case pkgconstant.TemplateOrderPaymentReminder:
		return s.generateOrderPaymentReminderEmail(payload)
	default:
		s.logger.Infof("Unknown template ID: %s", payload.TemplateID)

		return "", fmt.Errorf("unknown template ID: %s", payload.TemplateID)
	}
}

// generateOrderConfirmedEmail generates HTML email for order confirmation.
func (s *notificationEventService) generateOrderConfirmedEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order kafkaevent.OrderConfirmedData
	if err = json.Unmarshal(orderJSON, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order confirmation data: %w", err)
	}

	// Convert order items to template data
	var items []dto.OrderItemTemplateData
	for _, item := range order.Items {
		items = append(items, dto.OrderItemTemplateData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			TotalPrice:  item.TotalPrice.String(),
			Currency:    order.Currency,
		})
	}

	templateData := mapper.MapToOrderConfirmedTemplateData(
		order.CustomerName,
		order.OrderID.String(),
		order.OrderDate.Format("January 2, 2006"),
		order.CustomerEmail,
		items,
		order.Currency,
		order.Subtotal,
		order.ShippingCost,
		order.TotalTax,
		order.TotalDiscount,
		order.TotalPrice,
		order.TrackingNumber,
		order.EstimatedDelivery,
	)

	return s.emailService.RenderTemplate(constant.TemplateFileOrderConfirmed, templateData)
}

// generateOrderDeliveredEmail generates HTML email for order delivered notification.
func (s *notificationEventService) generateOrderDeliveredEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order kafkaevent.OrderConfirmedData
	if err = json.Unmarshal(orderJSON, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order confirmation data: %w", err)
	}

	// Convert order items to template data
	var items []dto.OrderItemTemplateData
	for _, item := range order.Items {
		items = append(items, dto.OrderItemTemplateData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			TotalPrice:  item.TotalPrice.String(),
			Currency:    order.Currency,
		})
	}

	templateData := mapper.MapToOrderDeliveredTemplateData(
		order.CustomerName,
		order.OrderID.String(),
		order.OrderDate.Format("January 2, 2006"),
		order.CustomerEmail,
		items,
		order.Currency,
		order.Subtotal,
		order.ShippingCost,
		order.TotalTax,
		order.TotalDiscount,
		order.TotalPrice,
		order.TrackingNumber,
		order.EstimatedDelivery,
		time.Now(),
	)

	return s.emailService.RenderTemplate(constant.TemplateFileOrderDelivered, templateData)
}

// generateOrderShippedEmail generates HTML email for order shipped notification.
func (s *notificationEventService) generateOrderShippedEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	orderNumber := ""
	trackingNumber := ""

	if orderData, exists := payload.Data["order_number"]; exists {
		if str, ok := orderData.(string); ok {
			orderNumber = str
		}
	}

	if trackingData, exists := payload.Data["tracking_number"]; exists {
		if str, ok := trackingData.(string); ok {
			trackingNumber = str
		}
	}

	templateData := &dto.OrderShippedTemplateData{
		RecipientName:  recipientName,
		OrderNumber:    orderNumber,
		TrackingNumber: trackingNumber,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileOrderShipped, templateData)
}

// generateOrderCancelledEmail generates HTML email for order cancellation.
func (s *notificationEventService) generateOrderCancelledEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	orderNumber := ""

	if orderData, exists := payload.Data["order_number"]; exists {
		if str, ok := orderData.(string); ok {
			orderNumber = str
		}
	}

	templateData := &dto.OrderCanceledTemplateData{
		RecipientName: recipientName,
		OrderNumber:   orderNumber,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileOrderCanceled, templateData)
}

// generateOrderPaymentRequiredEmail generates HTML email for payment required notification.
func (s *notificationEventService) generateOrderPaymentRequiredEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order kafkaevent.OrderConfirmedData
	if err = json.Unmarshal(orderJSON, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// Convert order items to template data
	var items []dto.OrderItemTemplateData
	for _, item := range order.Items {
		items = append(items, dto.OrderItemTemplateData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			TotalPrice:  item.TotalPrice.String(),
			Currency:    order.Currency,
		})
	}

	// Get payment deadline from payload data
	paymentDeadline := time.Now().Add(1 * time.Hour) // Default 1 hour

	if deadlineData, existsDeadline := payload.Data["payment_deadline"]; existsDeadline {
		if deadlineStr, ok := deadlineData.(string); ok {
			if parsedTime, errParse := time.Parse(time.RFC3339, deadlineStr); errParse == nil {
				paymentDeadline = parsedTime
			}
		}
	}

	// Get payment URL from payload data
	var paymentURL *string

	if urlData, existsURL := payload.Data["payment_url"]; existsURL {
		if urlStr, ok := urlData.(string); ok && urlStr != "" {
			paymentURL = &urlStr
		}
	}

	templateData := mapper.MapToOrderPaymentRequiredTemplateData(
		order.CustomerName,
		order.OrderID.String(),
		order.OrderDate.Format("January 2, 2006"),
		order.CustomerEmail,
		items,
		order.Currency,
		order.Subtotal,
		order.ShippingCost,
		order.TotalTax,
		order.TotalDiscount,
		order.TotalPrice,
		paymentDeadline,
		paymentURL,
	)

	return s.emailService.RenderTemplate(constant.TemplateFileOrderPaymentRequired, templateData)
}

// generateOrderPaymentReminderEmail generates HTML email for payment required notification.
func (s *notificationEventService) generateOrderPaymentReminderEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order kafkaevent.OrderConfirmedData
	if err = json.Unmarshal(orderJSON, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// Convert order items to template data
	var items []dto.OrderItemTemplateData
	for _, item := range order.Items {
		items = append(items, dto.OrderItemTemplateData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			TotalPrice:  item.TotalPrice.String(),
			Currency:    order.Currency,
		})
	}

	// Get payment deadline from payload data
	paymentDeadline := time.Now().Add(1 * time.Hour) // Default 1 hour

	if deadlineData, existsDeadline := payload.Data["payment_deadline"]; existsDeadline {
		if deadlineStr, ok := deadlineData.(string); ok {
			if parsedTime, errParse := time.Parse(time.RFC3339, deadlineStr); errParse == nil {
				paymentDeadline = parsedTime
			}
		}
	}

	// Get payment URL from payload data
	var paymentURL *string

	if urlData, existsURL := payload.Data["payment_url"]; existsURL {
		if urlStr, ok := urlData.(string); ok && urlStr != "" {
			paymentURL = &urlStr
		}
	}

	templateData := mapper.MapToPaymentReminderTemplateData(
		order.CustomerName,
		order.OrderID.String(),
		order.CustomerEmail,
		items,
		order.Currency,
		order.TotalPrice,
		paymentDeadline,
		paymentURL,
	)

	return s.emailService.RenderTemplate(constant.TemplateFileOrderPaymentReminder, templateData)
}

// generateOrderPaymentExpiredEmail generates HTML email for order payment expired notification.
func (s *notificationEventService) generateOrderPaymentExpiredEmail(
	payload *kafkaevent.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	orderNumber := ""
	orderDate := ""
	totalPrice := ""
	currency := ""

	// Extract order number from payload
	if orderData, exists := payload.Data["order_number"]; exists {
		if str, ok := orderData.(string); ok {
			orderNumber = str
		}
	}

	// Extract order date from payload
	if dateData, exists := payload.Data["order_date"]; exists {
		if str, ok := dateData.(string); ok {
			orderDate = str
		}
	}

	// Extract total price from payload
	if priceData, exists := payload.Data["total_price"]; exists {
		if str, ok := priceData.(string); ok {
			totalPrice = str
		}
	}

	// Extract currency from payload
	if currencyData, exists := payload.Data["currency"]; exists {
		if str, ok := currencyData.(string); ok {
			currency = str
		}
	}

	templateData := &dto.OrderPaymentExpiredTemplateData{
		RecipientName: recipientName,
		OrderNumber:   orderNumber,
		OrderDate:     orderDate,
		TotalPrice:    totalPrice,
		Currency:      currency,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileOrderPaymentExpired, templateData)
}

// sendPushNotification sends a push notification via SSE.
func (s *notificationEventService) sendPushNotification(
	ctx context.Context,
	payload *kafkaevent.NotificationRequestPayload,
) error {
	s.logger.Infof("Sending push notification to user %s with subject: %s",
		payload.RecipientUserID, payload.Subject)

	// Extract user ID
	userID, err := uuid.Parse(payload.RecipientUserID)
	if err != nil {
		return fmt.Errorf("invalid recipient user ID: %w", err)
	}

	// Use message if provided, otherwise use subject
	message := payload.Message
	if message == "" {
		message = payload.Subject
	}

	// Create push notification entity
	notif, err := entity.NewPushNotification(
		userID,
		constant.PushNotificationType(payload.NotificationType),
		payload.Subject,
		message,
		payload.Data,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification entity: %w", err)
	}

	// Save notification to database
	savedNotif, err := s.notificationRepo.Create(ctx, notif)
	if err != nil {
		return fmt.Errorf("failed to save notification to database: %w", err)
	}

	s.logger.Infof("Saved notification to database: %s", savedNotif.ID)

	// Map entity to DTO for proper JSON serialization
	notifDTO := mapper.MapToNotificationResponse(savedNotif)

	// Create SSE message with DTO (has proper json tags)
	sseMsg, err := sse.NewMessage(subscription.TypeNotificationCreated, notifDTO)
	if err != nil {
		return fmt.Errorf("failed to create SSE message: %w", err)
	}

	// Broadcast to local SSE connections
	if err = s.sseHub.BroadcastToUser(userID, sseMsg); err != nil {
		s.logger.Error("Failed to broadcast notification to local SSE connections",
			"user_id", userID,
			"error", err)
		// Don't return error - continue to publish to Redis
	}

	// Publish to Redis if event bus is available
	if s.eventBus != nil {
		if err = s.publishToRedis(ctx, userID, sseMsg); err != nil {
			s.logger.Error("Failed to publish notification to Redis",
				"user_id", userID,
				"error", err)
			// Don't return error - notification is already saved to DB
		}
	}

	s.logger.Infof("Successfully sent push notification to user %s", userID)

	return nil
}

// publishToRedis handles Redis publishing logic using sharded pub/sub.
func (s *notificationEventService) publishToRedis(
	ctx context.Context,
	userID uuid.UUID,
	sseMsg *sse.Message,
) error {
	redisEvent := &subscription.NotificationCreatedEvent{
		UserID:  userID,
		Message: sseMsg,
	}

	// Use user-based sharded channel for native Redis slot-based distribution
	channelName := redis.NotificationUserChannel(userID)

	baseEvent, err := rediseventbus.NewBaseEvent(
		s.instanceID,
		subscription.TypeNotificationCreated,
		redisEvent,
	)
	if err != nil {
		return fmt.Errorf("failed to create base event: %w", err)
	}

	// Use SPublish for sharded pub/sub (Redis 7.0+)
	if err = s.eventBus.SPublish(ctx, channelName, baseEvent); err != nil {
		return fmt.Errorf("failed to publish to sharded channel %s: %w", channelName, err)
	}

	s.logger.Debug("Published notification to Redis sharded channel",
		"user_id", userID,
		"channel", channelName)

	return nil
}

// CreateAndBroadcastNotification creates a notification and broadcasts it via SSE/Redis.
func (s *notificationEventService) CreateAndBroadcastNotification(
	ctx context.Context,
	userID *uuid.UUID,
	notificationType constant.PushNotificationType,
	title string,
	message string,
	metadata map[string]any,
) (*dto.NotificationResponse, error) {
	// Validate required fields
	if userID != nil {
		s.logger.Infof("Creating notification for user %s with title: %s", *userID, title)
	} else {
		s.logger.Infof("Creating broadcast notification with title: %s", title)
		return nil, errors.New("broadcast to all users not yet implemented")
	}

	// Create notification entity
	notif, err := entity.NewPushNotification(
		*userID,
		notificationType,
		title,
		message,
		metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification entity: %w", err)
	}

	// Save notification to database
	savedNotif, err := s.notificationRepo.Create(ctx, notif)
	if err != nil {
		return nil, fmt.Errorf("failed to save notification to database: %w", err)
	}

	s.logger.Infof("Saved notification to database: %s", savedNotif.ID)

	// Map entity to DTO for proper JSON serialization
	notifDTO := mapper.MapToNotificationResponse(savedNotif)

	// Create SSE message with DTO (has proper json tags)
	sseMsg, err := sse.NewMessage(subscription.TypeNotificationCreated, notifDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE message: %w", err)
	}

	// Broadcast to local SSE connections
	if err = s.sseHub.BroadcastToUser(*userID, sseMsg); err != nil {
		s.logger.Error("Failed to broadcast notification to local SSE connections",
			"user_id", *userID,
			"error", err)
		// Don't return error - continue to publish to Redis
	}

	// Publish to Redis if event bus is available
	if s.eventBus != nil {
		if err = s.publishToRedis(ctx, *userID, sseMsg); err != nil {
			s.logger.Error("Failed to publish notification to Redis",
				"user_id", *userID,
				"error", err)
			// Don't return error - notification is already saved to DB
		}
	}

	s.logger.Infof("Successfully created and broadcast notification to user %s", *userID)

	// Map to response DTO
	return mapper.MapToNotificationResponse(savedNotif), nil
}
