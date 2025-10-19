package models

// EmailMessage represents an email to be sent
type EmailMessage struct {
	To          []string          // Recipients
	From        string            // Sender email
	FromName    string            // Sender name
	Subject     string            // Email subject
	Body        string            // HTML body
	TextBody    string            // Plain text body
	ReplyTo     string            // Reply-to address
	CC          []string          // CC recipients
	BCC         []string          // BCC recipients
	Headers     []Header          // Custom headers
	MessageID   string            // Unique message ID
	Metadata    map[string]string // Additional metadata
}

// Header represents a custom email header
type Header struct {
	Name  string
	Value string
}

// EmailType represents the type of email being sent
type EmailType string

const (
	EmailTypeVerification  EmailType = "verification"
	EmailTypePasswordReset EmailType = "password_reset"
	EmailTypeWelcome       EmailType = "welcome"
	EmailTypeGeneral       EmailType = "general"
)
