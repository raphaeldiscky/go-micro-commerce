// Package smtputils provides utility functions for sending emails via SMTP.
package smtputils

import (
	"context"
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
)

// Mailer defines an interface for sending emails.
type Mailer interface {
	SendMail(ctx context.Context, to, subject, body string) error
}

// mailer is an implementation of the Mailer interface.
type mailer struct {
	dialer *gomail.Dialer
	cfg    *config.SMTPConfig
}

// NewMailer returns a singleton Mailer instance without using package-level globals.
func NewMailer(cfg *config.SMTPConfig) Mailer {
	var (
		once     sync.Once
		instance Mailer
		mu       sync.Mutex
	)

	return &struct{ Mailer }{
		Mailer: mailerFunc(func(ctx context.Context, to, subject, body string) error {
			mu.Lock()
			defer mu.Unlock()

			once.Do(func() {
				d := gomail.NewDialer(cfg.Host, cfg.Port, "", "")
				instance = &mailer{
					dialer: d,
					cfg:    cfg,
				}
			})

			return instance.SendMail(ctx, to, subject, body)
		}),
	}
}

// mailerFunc is an adapter to allow the use of ordinary functions as a Mailer.
type mailerFunc func(ctx context.Context, to, subject, body string) error

// SendMail sends an email using the provided function.
func (f mailerFunc) SendMail(ctx context.Context, to, subject, body string) error {
	return f(ctx, to, subject, body)
}

// SendMail sends an email using the provided function.
func (m *mailer) SendMail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := gomail.NewMessage()
		msg.SetHeader("From", m.cfg.Email)
		msg.SetHeader("To", to)
		msg.SetHeader("Subject", subject)
		msg.SetBody("text/html", body)

		if err := m.dialer.DialAndSend(msg); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		return nil
	}
}
