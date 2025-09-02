// Package mq provides the event definitions and handlers for the notification service.
package mq

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
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
	messageBody := fmt.Sprintf(constant.SendVerificationTemplate, verificationURL)

	if err := c.mailer.SendMail(ctx, evt.Payload.Email, subject, messageBody); err != nil {
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
	messageBody := fmt.Sprintf(constant.UserVerifiedTemplate)

	if err := c.mailer.SendMail(ctx, evt.Payload.Email, subject, messageBody); err != nil {
		c.logger.Errorf("failed to send email: %w", err)

		return err
	}

	c.logger.Infof("successfully verified email to %s", evt.Payload.Email)

	return nil
}
