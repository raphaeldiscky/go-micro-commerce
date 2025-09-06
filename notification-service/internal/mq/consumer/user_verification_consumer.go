// Package consumer provides the event definitions and handlers for the notification service.
package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/service"
)

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

// UserVerificationConsumer handles the logic for processing user verification requested events.
type UserVerificationConsumer struct {
	emailService service.EmailService
	logger       logger.Logger
}

// NewUserVerificationConsumer creates a new consumer for user verification requested events.
func NewUserVerificationConsumer(
	emailService service.EmailService,
	appLogger logger.Logger,
) *UserVerificationConsumer {
	return &UserVerificationConsumer{
		emailService: emailService,
		logger:       appLogger,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *UserVerificationConsumer) Handler(ctx context.Context, body []byte) error {
	var meta struct {
		Metadata event.Metadata `json:"metadata"`
	}

	if err := sonic.Unmarshal(body, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal event metadata: %w", err)
	}

	switch meta.Metadata.EventType {
	case constant.KafkaEventTypeEmailVerificationRequested:
		return c.handleVerificationRequested(ctx, body)
	case constant.KafkaEventTypeUserVerified:
		return c.handleUserVerified(ctx, body)
	default:
		c.logger.Warnf("ignoring event type: %s", meta.Metadata.EventType)

		return nil
	}
}

func (c *UserVerificationConsumer) handleVerificationRequested(
	ctx context.Context,
	body []byte,
) error {
	var evt EmailVerificationRequestedEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal EmailVerificationRequestedEvent: %w", err)
	}

	if evt.Payload.Email == "" {
		return fmt.Errorf("email is required but was empty")
	}

	if evt.Payload.Token == "" {
		return fmt.Errorf("token is required but was empty")
	}

	subject := constant.SendVerificationSubject
	verificationURL := fmt.Sprintf(
		"http://localhost:8080/auth/v1/verify?token=%s",
		evt.Payload.Token,
	)

	templateData := &dto.EmailVerificationTemplateData{
		RecipientName:   evt.Payload.Email,
		VerificationURL: verificationURL,
		TokenExpiresAt:  evt.Payload.TokenExpiresAt.Format(time.RFC3339),
	}

	// Use the template service to render the email
	messageBody, err := c.emailService.RenderTemplate(
		constant.TemplateFileEmailVerification,
		templateData,
	)
	if err != nil {
		return fmt.Errorf("failed to generate verification email body: %w", err)
	}

	if err := c.emailService.SendEmail(ctx, evt.Payload.Email, subject, messageBody); err != nil {
		return fmt.Errorf(
			"failed to send verification requested email to %s: %w",
			evt.Payload.Email,
			err,
		)
	}

	c.logger.Printf("successfully sent verification requested email to: %s", evt.Payload.Email)

	return nil
}

func (c *UserVerificationConsumer) handleUserVerified(ctx context.Context, body []byte) error {
	var evt UserVerifiedEvent
	if err := sonic.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal UserVerifiedEvent: %w", err)
	}

	// Send a confirmation email
	subject := constant.UserVerifiedSubject

	// Generate email body using template
	templateData := &dto.UserVerifiedTemplateData{
		RecipientName: evt.Payload.Email,
	}

	// Use the template service to render the email
	messageBody, err := c.emailService.RenderTemplate(
		constant.TemplateFileUserVerified,
		templateData,
	)
	if err != nil {
		return fmt.Errorf("failed to generate user verified email body: %w", err)
	}

	if err := c.emailService.SendEmail(ctx, evt.Payload.Email, subject, messageBody); err != nil {
		c.logger.Errorf("failed to send email: %w", err)

		return err
	}

	c.logger.Infof("successfully verified email to %s", evt.Payload.Email)

	return nil
}
