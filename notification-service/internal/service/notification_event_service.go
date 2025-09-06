// Package service provides business logic services for the notification service.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// ProcessEmailVerificationRequest handles email verification request events.
func (s *NotificationEventServiceImpl) ProcessEmailVerificationRequest(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing email verification request event: %s", inboxEvent.MessageID)

	var emailVerificationEvent EmailVerificationRequestedEvent
	if err := json.Unmarshal(inboxEvent.Payload, &emailVerificationEvent); err != nil {
		return fmt.Errorf("failed to unmarshal email verification event: %w", err)
	}

	subject := constant.SendVerificationSubject

	body, err := s.generateVerificationEmail(&emailVerificationEvent.Payload)
	if err != nil {
		return fmt.Errorf("failed to generate verification email: %w", err)
	}

	if err := s.emailService.SendEmail(ctx, emailVerificationEvent.Payload.Email, subject, body); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	s.logger.Infof(
		"Successfully sent verification email to: %s",
		emailVerificationEvent.Payload.Email,
	)

	return nil
}

// ProcessEmailUserVerified handles user verified events.
func (s *NotificationEventServiceImpl) ProcessEmailUserVerified(
	ctx context.Context,
	inboxEvent *entity.InboxEvent,
) error {
	s.logger.Infof("Processing user verified event: %s", inboxEvent.MessageID)

	var userVerifiedEvent UserVerifiedEvent
	if err := json.Unmarshal(inboxEvent.Payload, &userVerifiedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal user verified event: %w", err)
	}

	subject := constant.UserVerifiedSubject

	body, err := s.generateWelcomeEmail(&userVerifiedEvent.Payload)
	if err != nil {
		return fmt.Errorf("failed to generate welcome email: %w", err)
	}

	if err := s.emailService.SendEmail(ctx, userVerifiedEvent.Payload.Email, subject, body); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	s.logger.Infof("Successfully sent welcome email to: %s", userVerifiedEvent.Payload.Email)

	return nil
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

	if err := s.emailService.SendEmail(ctx, payload.RecipientEmail, payload.Subject, emailBody); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", payload.RecipientEmail, err)
	}

	s.logger.Infof("Successfully sent email to %s", payload.RecipientEmail)
	log.Printf("📧 EMAIL SENT: To=%s, Subject=%s, Template=%s",
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
	case pkgconstant.TemplateOrderConfirmation:
		return s.generateOrderConfirmationEmail(payload)
	case pkgconstant.TemplateOrderShipped:
		return s.generateOrderShippedEmail(payload)
	case pkgconstant.TemplateOrderCanceled:
		return s.generateOrderCancelledEmail(payload)
	case pkgconstant.TemplatePaymentConfirmation:
		return s.generatePaymentConfirmationEmail(payload)
	default:
		return s.generateGenericEmail(payload)
	}
}

// generateOrderConfirmationEmail generates HTML email for order confirmation.
func (s *NotificationEventServiceImpl) generateOrderConfirmationEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", fmt.Errorf("order data not found in payload")
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %w", err)
	}

	var order event.OrderConfirmationData
	if err := json.Unmarshal(orderJSON, &order); err != nil {
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

	templateData := mapper.MapToOrderConfirmationTemplateData(
		order.CustomerName,
		order.OrderNumber,
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

	return s.emailService.RenderTemplate(constant.TemplateFileOrderConfirmation, templateData)
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

// generatePaymentConfirmationEmail generates HTML email for payment confirmation.
func (s *NotificationEventServiceImpl) generatePaymentConfirmationEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	amount := ""
	paymentMethod := ""

	if amountData, exists := payload.Data["amount"]; exists {
		if str, ok := amountData.(string); ok {
			amount = str
		}
	}

	if methodData, exists := payload.Data["payment_method"]; exists {
		if str, ok := methodData.(string); ok {
			paymentMethod = str
		}
	}

	templateData := &dto.PaymentConfirmationTemplateData{
		RecipientName: recipientName,
		Amount:        amount,
		PaymentMethod: paymentMethod,
	}

	return s.emailService.RenderTemplate(constant.TemplateFilePaymentConfirmation, templateData)
}

// generateGenericEmail generates a generic HTML email template.
func (s *NotificationEventServiceImpl) generateGenericEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	templateData := &dto.GenericTemplateData{
		Subject:       payload.Subject,
		RecipientName: recipientName,
		Data:          payload.Data,
	}

	return s.emailService.RenderTemplate(constant.TemplateFileGeneric, templateData)
}
