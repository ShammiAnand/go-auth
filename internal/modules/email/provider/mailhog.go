package provider

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/shammianand/go-auth/internal/modules/email/models"
)

// MailhogProvider implements EmailProvider for local testing with Mailhog
type MailhogProvider struct {
	smtpHost    string
	smtpPort    string
	defaultFrom string
	logger      *slog.Logger
}

// NewMailhogProvider creates a new Mailhog provider
func NewMailhogProvider(smtpHost, smtpPort, defaultFrom string, logger *slog.Logger) EmailProvider {
	if logger == nil {
		logger = slog.Default()
	}

	return &MailhogProvider{
		smtpHost:    smtpHost,
		smtpPort:    smtpPort,
		defaultFrom: defaultFrom,
		logger:      logger,
	}
}

// GetProviderName returns the provider name
func (mp *MailhogProvider) GetProviderName() string {
	return "mailhog"
}

// SendEmail sends a single email message
func (mp *MailhogProvider) SendEmail(msg *models.EmailMessage) error {
	mp.logger.Info("Sending email via Mailhog",
		"to", strings.Join(msg.To, ", "),
		"subject", msg.Subject,
		"from", msg.From,
	)

	// Use default sender if not specified
	if msg.From == "" {
		msg.From = mp.defaultFrom
	}

	// Build email content
	var emailContent strings.Builder

	// From header
	if msg.FromName != "" {
		emailContent.WriteString(fmt.Sprintf("From: %s <%s>\r\n", msg.FromName, msg.From))
	} else {
		emailContent.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	}

	// To header
	emailContent.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// CC header
	if len(msg.CC) > 0 {
		emailContent.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}

	// Reply-To header
	if msg.ReplyTo != "" {
		emailContent.WriteString(fmt.Sprintf("Reply-To: %s\r\n", msg.ReplyTo))
	}

	// Subject
	emailContent.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// Date
	emailContent.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	// Message ID
	if msg.MessageID != "" {
		emailContent.WriteString(fmt.Sprintf("Message-ID: <%s>\r\n", msg.MessageID))
	}

	// Custom headers
	for _, header := range msg.Headers {
		emailContent.WriteString(fmt.Sprintf("%s: %s\r\n", header.Name, header.Value))
	}

	// MIME headers
	emailContent.WriteString("MIME-Version: 1.0\r\n")

	// Handle multipart if we have both HTML and text
	if msg.Body != "" && msg.TextBody != "" {
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		emailContent.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		emailContent.WriteString("\r\n")

		// Text part
		emailContent.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		emailContent.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		emailContent.WriteString("\r\n")
		emailContent.WriteString(msg.TextBody)
		emailContent.WriteString("\r\n")

		// HTML part
		emailContent.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		emailContent.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		emailContent.WriteString("\r\n")
		emailContent.WriteString(msg.Body)
		emailContent.WriteString("\r\n")

		// End boundary
		emailContent.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.Body != "" {
		// HTML only
		emailContent.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		emailContent.WriteString("\r\n")
		emailContent.WriteString(msg.Body)
	} else {
		// Text only
		emailContent.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		emailContent.WriteString("\r\n")
		emailContent.WriteString(msg.TextBody)
	}

	// Send via SMTP
	addr := fmt.Sprintf("%s:%s", mp.smtpHost, mp.smtpPort)

	// Mailhog doesn't require authentication
	err := smtp.SendMail(
		addr,
		nil, // No auth needed for Mailhog
		msg.From,
		append(append(msg.To, msg.CC...), msg.BCC...),
		[]byte(emailContent.String()),
	)

	if err != nil {
		mp.logger.Error("Failed to send email via Mailhog", "error", err)
		return fmt.Errorf("mailhog send failed: %w", err)
	}

	mp.logger.Info("Email sent successfully via Mailhog",
		"to", strings.Join(msg.To, ", "),
		"messageId", msg.MessageID,
	)

	return nil
}

// SendBatch sends multiple emails
func (mp *MailhogProvider) SendBatch(messages []*models.EmailMessage) error {
	for _, msg := range messages {
		if err := mp.SendEmail(msg); err != nil {
			return fmt.Errorf("batch send failed: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
