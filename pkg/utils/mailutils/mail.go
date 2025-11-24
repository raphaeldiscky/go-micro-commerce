// Package mailutils provides utility functions for sending emails via MailHog or SendGrid.
package mailutils

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/gomail.v2"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
)

// ErrInvalidMailProvider is returned when an invalid mail provider is specified.
var ErrInvalidMailProvider = errors.New("invalid mail provider, must be 'mailhog' or 'sendgrid'")

// Mailer defines an interface for sending emails.
type Mailer interface {
	SendMail(ctx context.Context, to, subject, body string) error
}

// mailhogMailer is a MailHog/SMTP implementation of the Mailer interface (for local development).
type mailhogMailer struct {
	dialer *gomail.Dialer
	cfg    *config.MailConfig
}

// sendgridMailer is a SendGrid implementation of the Mailer interface (for production).
type sendgridMailer struct {
	client *sendgrid.Client
	cfg    *config.MailConfig
}

// NewMailer returns a Mailer instance based on the configured provider.
// Provider "sendgrid" uses SendGrid API, provider "mailhog" uses SMTP (MailHog for local).
func NewMailer(cfg *config.MailConfig) (Mailer, error) {
	switch cfg.Provider {
	case "sendgrid":
		if cfg.SendGridAPIKey == "" {
			return nil, errors.New("SendGrid API key is required for sendgrid provider")
		}

		return &sendgridMailer{
			client: sendgrid.NewSendClient(cfg.SendGridAPIKey),
			cfg:    cfg,
		}, nil
	case "mailhog", "":
		dialer := gomail.NewDialer(cfg.Host, cfg.Port, "", "")

		return &mailhogMailer{
			dialer: dialer,
			cfg:    cfg,
		}, nil
	default:
		return nil, ErrInvalidMailProvider
	}
}

// SendMail sends an email using SMTP (for MailHog/local development).
func (m *mailhogMailer) SendMail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := gomail.NewMessage()
		msg.SetHeader("From", m.cfg.FromEmail)
		msg.SetHeader("To", to)
		msg.SetHeader("Subject", subject)
		msg.SetBody("text/html", body)

		if err := m.dialer.DialAndSend(msg); err != nil {
			return fmt.Errorf("failed to send email via MailHog: %w", err)
		}

		return nil
	}
}

// SendMail sends an email using SendGrid API (for production).
func (s *sendgridMailer) SendMail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		from := mail.NewEmail("", s.cfg.FromEmail)
		toEmail := mail.NewEmail("", to)
		message := mail.NewSingleEmail(from, subject, toEmail, "", body)

		response, err := s.client.SendWithContext(ctx, message)
		if err != nil {
			return fmt.Errorf("failed to send email via SendGrid: %w", err)
		}

		// SendGrid returns 2xx status codes for success
		if response.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf(
				"SendGrid returned error status %d: %s",
				response.StatusCode,
				response.Body,
			)
		}

		return nil
	}
}
