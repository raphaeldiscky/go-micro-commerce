// Package service provides business logic services for the notification service.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// NotificationRequestEvent is the envelope for notification request events.
type NotificationRequestEvent struct {
	Metadata event.Metadata                   `json:"metadata"`
	Payload  event.NotificationRequestPayload `json:"payload"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata event.Metadata            `json:"metadata"`
	Payload  event.UserVerifiedPayload `json:"payload"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata event.Metadata                          `json:"metadata"`
	Payload  event.EmailVerificationRequestedPayload `json:"payload"`
}

// NotificationEventService handles all notification business logic.
type NotificationEventService interface {
	ProcessNotificationRequest(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessEmailVerificationRequest(ctx context.Context, inboxEvent *entity.InboxEvent) error
	ProcessEmailUserVerified(ctx context.Context, inboxEvent *entity.InboxEvent) error
}

// NotificationEventServiceImpl implements notification business logic.
type NotificationEventServiceImpl struct {
	dataStore    repository.DataStore
	emailService EmailService
	logger       logger.Logger
}

// NewNotificationEventService creates a new notification service instance.
func NewNotificationEventService(
	dataStore repository.DataStore,
	emailService EmailService,
	appLogger logger.Logger,
) NotificationEventService {
	return &NotificationEventServiceImpl{
		dataStore:    dataStore,
		emailService: emailService,
		logger:       appLogger,
	}
}

// ProcessNotificationRequest handles notification request events.
func (s *NotificationEventServiceImpl) ProcessNotificationRequest(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing notification request event: %s", inboxEvent.MessageID)

	var notificationEvent NotificationRequestEvent
	if err := json.Unmarshal(inboxEvent.Payload, &notificationEvent); err != nil {
		return fmt.Errorf("failed to unmarshal notification request event: %w", err)
	}

	switch notificationEvent.Payload.NotificationType {
	case event.NotificationTypeEmail:
		return s.sendEmailNotification(ctx, &notificationEvent.Payload)
	case event.NotificationTypeSMS:
		s.logger.Info("SMS notifications not yet implemented")

		return nil
	case event.NotificationTypePush:
		s.logger.Info("Push notifications not yet implemented")

		return nil
	default:
		return fmt.Errorf(
			"unsupported notification type: %s",
			notificationEvent.Payload.NotificationType,
		)
	}
}

// processEmailEvent is a generic helper for processing email events.
func (s *NotificationEventServiceImpl) processEmailEvent(
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
	case "email verification request":
		return constant.SendVerificationSubject
	case "user verified":
		return constant.UserVerifiedSubject
	default:
		return ""
	}
}

// ProcessEmailVerificationRequest handles email verification request events.
func (s *NotificationEventServiceImpl) ProcessEmailVerificationRequest(
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

	return s.processEmailEvent(ctx, inboxEvent, "email verification request", unmarshalFn)
}

// ProcessEmailUserVerified handles user verified events.
func (s *NotificationEventServiceImpl) ProcessEmailUserVerified(
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

	return s.processEmailEvent(ctx, inboxEvent, "user verified", unmarshalFn)
}

// generateVerificationEmail creates an email verification email body.
func (s *NotificationEventServiceImpl) generateVerificationEmail(
	payload *event.EmailVerificationRequestedPayload,
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
func (s *NotificationEventServiceImpl) generateWelcomeEmail(
	payload *event.UserVerifiedPayload,
) (string, error) {
	templateData := &dto.UserVerifiedTemplateData{
		RecipientName: payload.Email,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileUserVerified, templateData)
}

// sendEmailNotification sends an email notification.
func (s *NotificationEventServiceImpl) sendEmailNotification(
	ctx context.Context,
	payload *event.NotificationRequestPayload,
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
func (s *NotificationEventServiceImpl) generateEmailBody(
	payload *event.NotificationRequestPayload,
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
func (s *NotificationEventServiceImpl) generateOrderConfirmedEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order event.OrderConfirmedData
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
func (s *NotificationEventServiceImpl) generateOrderDeliveredEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order event.OrderConfirmedData
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
func (s *NotificationEventServiceImpl) generateOrderShippedEmail(
	payload *event.NotificationRequestPayload,
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
func (s *NotificationEventServiceImpl) generateOrderCancelledEmail(
	payload *event.NotificationRequestPayload,
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
func (s *NotificationEventServiceImpl) generateOrderPaymentRequiredEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order event.OrderConfirmedData
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
func (s *NotificationEventServiceImpl) generateOrderPaymentReminderEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", errors.New("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order event.OrderConfirmedData
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
