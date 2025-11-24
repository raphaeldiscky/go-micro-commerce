// Package service provides business logic services for the notification service.
package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/mailutils"
)

// EmailService handles email template rendering and sending operations.
type EmailService interface {
	LoadTemplate(templateName string) (string, error)
	RenderTemplate(templateName string, data any) (string, error)
	SendEmail(ctx context.Context, to, subject, body string) error
}

// emailService handles loading and rendering email templates.
type emailService struct {
	templatesPath string
	mailer        mailutils.Mailer
}

// NewEmailService creates a new template service instance.
func NewEmailService(templatesPath string, mailer mailutils.Mailer) EmailService {
	return &emailService{
		templatesPath: templatesPath,
		mailer:        mailer,
	}
}

// LoadTemplate loads a template file and returns its content.
func (ts *emailService) LoadTemplate(templateName string) (string, error) {
	// Validate template name to prevent directory traversal
	if filepath.Base(templateName) != templateName {
		return "", fmt.Errorf("invalid template name: %s", templateName)
	}

	templatePath := filepath.Join(ts.templatesPath, templateName)

	// Ensure the resolved path is still within the templates directory
	resolvedPath, err := filepath.Abs(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve template path: %w", err)
	}

	absTemplatesPath, err := filepath.Abs(ts.templatesPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve templates directory: %w", err)
	}

	// Use proper path checking instead of deprecated filepath.HasPrefix
	relPath, err := filepath.Rel(absTemplatesPath, resolvedPath)
	if err != nil || strings.HasPrefix(relPath, "..") ||
		strings.Contains(relPath, string(filepath.Separator)+"..") {
		return "", fmt.Errorf("template path outside of templates directory: %s", templateName)
	}

	// Check if file exists
	if _, err = os.Stat(resolvedPath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file %s not found: %w", resolvedPath, err)
	}

	// Read template content
	// #nosec G304 - Path is validated above to prevent directory traversal
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", resolvedPath, err)
	}

	return string(content), nil
}

// RenderTemplate renders a template with the provided data using html/template.
func (ts *emailService) RenderTemplate(templateName string, data any) (string, error) {
	templateContent, err := ts.LoadTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", templateName, err)
	}

	// Parse the template
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Render the template with data
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// SendEmail sends an email using the configured mailer.
func (ts *emailService) SendEmail(ctx context.Context, to, subject, body string) error {
	return ts.mailer.SendMail(ctx, to, subject, body)
}
