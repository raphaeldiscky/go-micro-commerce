package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

// NotificationRequestEvent is the envelope for notification request events.
type NotificationRequestEvent struct {
	Metadata event.Metadata                   `json:"metadata"`
	Payload  event.NotificationRequestPayload `json:"payload"`
}

// NotificationRequestConsumer handles notification request events from other services.
type NotificationRequestConsumer struct {
	logger       logger.Logger
	emailService service.EmailService
}

// NewNotificationRequestConsumer creates a new consumer for notification request events.
func NewNotificationRequestConsumer(
	emailService service.EmailService,
	appLogger logger.Logger,
) *NotificationRequestConsumer {
	return &NotificationRequestConsumer{
		emailService: emailService,
		logger:       appLogger,
	}
}

// Handler processes notification request events.
func (c *NotificationRequestConsumer) Handler(ctx context.Context, body []byte) error {
	// First, extract metadata to understand the event
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	c.logger.Infof("Processing notification request event: %s from %s",
		meta.Metadata.EventID, meta.Metadata.Source)

	// Process the notification request based on event type
	switch meta.Metadata.EventType {
	case kafka.NotificationRequestedEventType:
		return c.processNotificationRequest(ctx, body)
	default:
		c.logger.Warnf("ignoring unknown notification event type: %s", meta.Metadata.EventType)

		return nil
	}
}

// processNotificationRequest handles notification request events to send emails/notifications.
func (c *NotificationRequestConsumer) processNotificationRequest(
	ctx context.Context,
	body []byte,
) error {
	var evt NotificationRequestEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal notification request event: %w", err)
	}

	c.logger.Infof("Processing notification request: Type=%s, To=%s, Subject=%s",
		evt.Payload.NotificationType,
		evt.Payload.RecipientEmail,
		evt.Payload.Subject)

	// Handle different notification types
	switch evt.Payload.NotificationType {
	case event.NotificationTypeEmail:
		return c.sendEmail(ctx, &evt.Payload)
	case event.NotificationTypeSMS:
		return c.sendSMS(ctx, &evt.Payload)
	case event.NotificationTypePush:
		return c.sendPushNotification(ctx, &evt.Payload)
	default:
		return fmt.Errorf("unsupported notification type: %s", evt.Payload.NotificationType)
	}
}

// sendEmail sends an email notification.
func (c *NotificationRequestConsumer) sendEmail(
	ctx context.Context,
	payload *event.NotificationRequestPayload,
) error {
	c.logger.Infof("Sending email to %s with subject: %s",
		payload.RecipientEmail, payload.Subject)

	// Generate email body based on template
	emailBody, err := c.generateEmailBody(payload)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %w", err)
	}

	// Send email using SMTP mailer
	if err := c.emailService.SendEmail(ctx, payload.RecipientEmail, payload.Subject, emailBody); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", payload.RecipientEmail, err)
	}

	c.logger.Infof("Successfully sent email to %s", payload.RecipientEmail)
	log.Printf("📧 EMAIL SENT: To=%s, Subject=%s, Template=%s",
		payload.RecipientEmail, payload.Subject, payload.TemplateID)

	return nil
}

// generateEmailBody generates the email body based on template ID and data.
func (c *NotificationRequestConsumer) generateEmailBody(
	payload *event.NotificationRequestPayload,
) (string, error) {
	c.logger.Infof(
		"Processing template ID: '%s' for email: %s",
		payload.TemplateID,
		payload.RecipientEmail,
	)

	switch payload.TemplateID {
	case pkgconstant.TemplateOrderConfirmation:
		return c.generateOrderConfirmationEmail(payload)
	case pkgconstant.TemplateOrderShipped:
		return c.generateOrderShippedEmail(payload)
	case pkgconstant.TemplateOrderCanceled:
		return c.generateOrderCancelledEmail(payload)
	case pkgconstant.TemplatePaymentConfirmation:
		return c.generatePaymentConfirmationEmail(payload)
	default:
		// Generic template - just include the data as a simple HTML format
		return c.generateGenericEmail(payload)
	}
}

// generateOrderConfirmationEmail generates HTML email for order confirmation.
func (c *NotificationRequestConsumer) generateOrderConfirmationEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	orderData, exists := payload.Data["order"]
	if !exists {
		return "", fmt.Errorf("order data not found in payload")
	}

	// Convert to JSON for easy access
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

	// Use the template service to render the email
	return c.emailService.RenderTemplate(constant.TemplateFileOrderConfirmation, templateData)
}

// generateOrderShippedEmail generates HTML email for order shipped notification.
func (c *NotificationRequestConsumer) generateOrderShippedEmail(
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

	// Create template data struct
	templateData := &dto.OrderShippedTemplateData{
		RecipientName:  recipientName,
		OrderNumber:    orderNumber,
		TrackingNumber: trackingNumber,
	}

	// Use the template service to render the email
	return c.emailService.RenderTemplate(constant.TemplateFileOrderShipped, templateData)
}

// generateOrderCancelledEmail generates HTML email for order cancellation.
func (c *NotificationRequestConsumer) generateOrderCancelledEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	orderNumber := ""

	if orderData, exists := payload.Data["order_number"]; exists {
		if str, ok := orderData.(string); ok {
			orderNumber = str
		}
	}

	// Create template data struct
	templateData := &dto.OrderCanceledTemplateData{
		RecipientName: recipientName,
		OrderNumber:   orderNumber,
	}

	// Use the template service to render the email
	return c.emailService.RenderTemplate(constant.TemplateFileOrderCanceled, templateData)
}

// generatePaymentConfirmationEmail generates HTML email for payment confirmation.
func (c *NotificationRequestConsumer) generatePaymentConfirmationEmail(
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

	// Create template data struct
	templateData := &dto.PaymentConfirmationTemplateData{
		RecipientName: recipientName,
		Amount:        amount,
		PaymentMethod: paymentMethod,
	}

	// Use the template service to render the email
	return c.emailService.RenderTemplate(constant.TemplateFilePaymentConfirmation, templateData)
}

// generateGenericEmail generates a generic HTML email template.
func (c *NotificationRequestConsumer) generateGenericEmail(
	payload *event.NotificationRequestPayload,
) (string, error) {
	recipientName := payload.RecipientName
	// Create template data struct
	templateData := &dto.GenericTemplateData{
		Subject:       payload.Subject,
		RecipientName: recipientName,
		Data:          payload.Data,
	}

	// Use the template service to render the email
	return c.emailService.RenderTemplate(constant.TemplateFileGeneric, templateData)
}

// sendSMS sends an SMS notification.
func (c *NotificationRequestConsumer) sendSMS(
	_ context.Context,
	_ *event.NotificationRequestPayload,
) error {
	c.logger.Infof("SMS notifications not yet implemented")

	return nil
}

// sendPushNotification sends a push notification.
func (c *NotificationRequestConsumer) sendPushNotification(
	_ context.Context,
	_ *event.NotificationRequestPayload,
) error {
	c.logger.Infof("Push notifications not yet implemented")

	return nil
}
