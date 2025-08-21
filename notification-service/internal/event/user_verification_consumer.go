// Package event provides the event definitions and handlers for the notification service.
package event

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/constant"
)

// EmailVerificationRequestedPayload holds the data for the email verification requested event.
type EmailVerificationRequestedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

// EmailVerificationRequestedEvent is the envelope for all email verification requested events.
type EmailVerificationRequestedEvent struct {
	Metadata mq.KafkaMetadata                  `json:"metadata"`
	Payload  EmailVerificationRequestedPayload `json:"payload"`
}

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	Metadata mq.KafkaMetadata    `json:"metadata"`
	Payload  UserVerifiedPayload `json:"payload"`
}

// UserVerificationConsumer handles the logic for processing user verification requested events.
type UserVerificationConsumer struct {
	mailer smtputils.Mailer
	logger logger.Logger
}

// NewUserVerificationConsumer creates a new consumer for user verification requested events.
func NewUserVerificationConsumer(
	mailer smtputils.Mailer,
	appLogger logger.Logger,
) *UserVerificationConsumer {
	return &UserVerificationConsumer{
		mailer: mailer,
		logger: appLogger,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *UserVerificationConsumer) Handler(ctx context.Context, body []byte) error {
	var meta struct {
		Metadata mq.KafkaMetadata `json:"metadata"`
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
		return fmt.Errorf("unknown event type: %s", meta.Metadata.EventType)
	}
}

func (c *UserVerificationConsumer) handleVerificationRequested(
	ctx context.Context,
	body []byte,
) error {
	var event EmailVerificationRequestedEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal EmailVerificationRequestedEvent: %w", err)
	}

	if event.Payload.Email == "" {
		return fmt.Errorf("email is required but was empty")
	}

	if event.Payload.Token == "" {
		return fmt.Errorf("token is required but was empty")
	}

	subject := constant.SendVerificationSubject
	verificationURL := fmt.Sprintf(
		"http://localhost:8080/auth/v1/verify?token=%s",
		event.Payload.Token,
	)
	messageBody := fmt.Sprintf(constant.SendVerificationTemplate, verificationURL)

	if err := c.mailer.SendMail(ctx, event.Payload.Email, subject, messageBody); err != nil {
		return fmt.Errorf(
			"failed to send verification requested email to %s: %w",
			event.Payload.Email,
			err,
		)
	}

	c.logger.Printf("successfully sent verification requested email to: %s", event.Payload.Email)

	return nil
}

func (c *UserVerificationConsumer) handleUserVerified(ctx context.Context, body []byte) error {
	var event UserVerifiedEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal UserVerifiedEvent: %w", err)
	}

	// Send a confirmation email
	subject := constant.UserVerifiedSubject
	messageBody := fmt.Sprintf(constant.UserVerifiedTemplate)

	if err := c.mailer.SendMail(ctx, event.Payload.Email, subject, messageBody); err != nil {
		c.logger.Errorf("failed to send email: %w", err)

		return err
	}

	c.logger.Infof("successfully verified email to %s", event.Payload.Email)

	return nil
}
