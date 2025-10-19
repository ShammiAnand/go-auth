package provider

import (
	"fmt"
	"log/slog"

	"github.com/shammianand/go-auth/internal/modules/email/models"
)

// SESProvider implements EmailProvider for AWS SES (stub for future implementation)
type SESProvider struct {
	apiKey    string
	secretKey string
	region    string
	logger    *slog.Logger
}

// NewSESProvider creates a new AWS SES provider (stub)
func NewSESProvider(apiKey, secretKey, region string, logger *slog.Logger) EmailProvider {
	if logger == nil {
		logger = slog.Default()
	}

	return &SESProvider{
		apiKey:    apiKey,
		secretKey: secretKey,
		region:    region,
		logger:    logger,
	}
}

// GetProviderName returns the provider name
func (sp *SESProvider) GetProviderName() string {
	return "aws_ses"
}

// SendEmail sends a single email via AWS SES (stub implementation)
func (sp *SESProvider) SendEmail(msg *models.EmailMessage) error {
	sp.logger.Warn("SES provider is not yet implemented - email not sent",
		"to", msg.To,
		"subject", msg.Subject,
	)

	// TODO: Implement AWS SES integration
	// 1. Initialize AWS SDK session
	// 2. Create SES client
	// 3. Build raw email message
	// 4. Send via SendRawEmail API
	// 5. Handle response and errors

	return fmt.Errorf("SES provider not yet implemented")
}

// SendBatch sends multiple emails (stub implementation)
func (sp *SESProvider) SendBatch(messages []*models.EmailMessage) error {
	sp.logger.Warn("SES batch send not yet implemented")

	// TODO: Implement batch sending
	// Consider using SES SendBulkTemplatedEmail for efficiency

	return fmt.Errorf("SES batch send not yet implemented")
}
