package event

import (
	"context"
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/smtputils"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/constant"
)

// EmailVerificationRequestedEvent represents the email verification event payload.
type EmailVerificationRequestedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// EmailVerificationConsumer handles the logic for processing email verification events.
type EmailVerificationConsumer struct {
	mailer smtputils.Mailer
}

// NewEmailVerificationConsumer creates a new consumer for email verification events.
// It requires a mailer to be injected as a dependency.
func NewEmailVerificationConsumer(mailer smtputils.Mailer) *EmailVerificationConsumer {
	return &EmailVerificationConsumer{
		mailer: mailer,
	}
}

// Handler is the method that implements mq.KafkaHandler. It contains the business logic.
func (c *EmailVerificationConsumer) Handler(ctx context.Context, body []byte) error {
	var event EmailVerificationRequestedEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		log.Printf("failed to unmarshal EmailVerificationRequestedEvent: %s", err)

		return nil
	}

	log.Printf("Processing email verification for: %s", event.Email)

	// Here you would have more complex logic, e.g., fetching a template
	subject := constant.SendVerificationSubject
	messageBody := fmt.Sprintf(constant.SendVerificationTemplate, event.Token)

	// Use the injected mailer to send the email
	return c.mailer.SendMail(ctx, event.Email, subject, messageBody)
}
