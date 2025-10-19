package provider

import "github.com/shammianand/go-auth/internal/modules/email/models"

// EmailProvider defines the interface for email service providers
type EmailProvider interface {
	// GetProviderName returns the name of the provider
	GetProviderName() string

	// SendEmail sends a single email
	SendEmail(msg *models.EmailMessage) error

	// SendBatch sends multiple emails
	SendBatch(messages []*models.EmailMessage) error
}
