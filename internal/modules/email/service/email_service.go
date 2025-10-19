package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/internal/modules/email/models"
	"github.com/shammianand/go-auth/internal/modules/email/provider"
)

// EmailService handles email operations
type EmailService struct {
	provider provider.EmailProvider
	client   *ent.Client
	logger   *slog.Logger
	fromEmail string
	fromName  string
}

// NewEmailService creates a new email service
func NewEmailService(provider provider.EmailProvider, client *ent.Client, logger *slog.Logger, fromEmail, fromName string) *EmailService {
	if logger == nil {
		logger = slog.Default()
	}

	return &EmailService{
		provider:  provider,
		client:    client,
		logger:    logger,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendVerificationEmail sends an email verification link
func (s *EmailService) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email, firstName, token string) error {
	// Generate verification link
	verificationLink := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)

	msg := &models.EmailMessage{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  "Verify your email address",
		Body: s.buildVerificationHTML(firstName, verificationLink),
		TextBody: s.buildVerificationText(firstName, verificationLink),
		MessageID: fmt.Sprintf("%s@go-auth", uuid.New().String()),
		Metadata: map[string]string{
			"user_id": userID.String(),
			"type":    string(models.EmailTypeVerification),
		},
	}

	// Send email
	err := s.provider.SendEmail(msg)

	// Log email delivery
	status := "sent"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	_, logErr := s.client.EmailLogs.Create().
		SetUserID(userID).
		SetRecipient(email).
		SetEmailType(string(models.EmailTypeVerification)).
		SetSubject(msg.Subject).
		SetStatus(status).
		SetProvider(s.provider.GetProviderName()).
		SetProviderMessageID(msg.MessageID).
		SetNillableErrorMessage(&errMsg).
		Save(ctx)

	if logErr != nil {
		s.logger.Error("Failed to log email", "error", logErr)
	}

	return err
}

// SendPasswordResetEmail sends a password reset link
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, userID uuid.UUID, email, firstName, token string) error {
	// Generate reset link
	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	msg := &models.EmailMessage{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  "Reset your password",
		Body:     s.buildPasswordResetHTML(firstName, resetLink),
		TextBody: s.buildPasswordResetText(firstName, resetLink),
		MessageID: fmt.Sprintf("%s@go-auth", uuid.New().String()),
		Metadata: map[string]string{
			"user_id": userID.String(),
			"type":    string(models.EmailTypePasswordReset),
		},
	}

	// Send email
	err := s.provider.SendEmail(msg)

	// Log email delivery
	status := "sent"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	_, logErr := s.client.EmailLogs.Create().
		SetUserID(userID).
		SetRecipient(email).
		SetEmailType(string(models.EmailTypePasswordReset)).
		SetSubject(msg.Subject).
		SetStatus(status).
		SetProvider(s.provider.GetProviderName()).
		SetProviderMessageID(msg.MessageID).
		SetNillableErrorMessage(&errMsg).
		Save(ctx)

	if logErr != nil {
		s.logger.Error("Failed to log email", "error", logErr)
	}

	return err
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, userID uuid.UUID, email, firstName string) error {
	msg := &models.EmailMessage{
		To:       []string{email},
		From:     s.fromEmail,
		FromName: s.fromName,
		Subject:  "Welcome to Go-Auth!",
		Body:     s.buildWelcomeHTML(firstName),
		TextBody: s.buildWelcomeText(firstName),
		MessageID: fmt.Sprintf("%s@go-auth", uuid.New().String()),
		Metadata: map[string]string{
			"user_id": userID.String(),
			"type":    string(models.EmailTypeWelcome),
		},
	}

	// Send email
	err := s.provider.SendEmail(msg)

	// Log email delivery
	status := "sent"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	_, logErr := s.client.EmailLogs.Create().
		SetUserID(userID).
		SetRecipient(email).
		SetEmailType(string(models.EmailTypeWelcome)).
		SetSubject(msg.Subject).
		SetStatus(status).
		SetProvider(s.provider.GetProviderName()).
		SetProviderMessageID(msg.MessageID).
		SetNillableErrorMessage(&errMsg).
		Save(ctx)

	if logErr != nil {
		s.logger.Error("Failed to log email", "error", logErr)
	}

	return err
}

// Template builders

func (s *EmailService) buildVerificationHTML(firstName, link string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Verify Your Email Address</h2>
        <p>Hi %s,</p>
        <p>Thank you for signing up! Please verify your email address by clicking the button below:</p>
        <div style="margin: 30px 0;">
            <a href="%s" style="background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Verify Email</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        <p>This link will expire in 24 hours.</p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="font-size: 12px; color: #999;">This is an automated message from Go-Auth.</p>
    </div>
</body>
</html>
`, firstName, link, link)
}

func (s *EmailService) buildVerificationText(firstName, link string) string {
	return fmt.Sprintf(`
Verify Your Email Address

Hi %s,

Thank you for signing up! Please verify your email address by visiting:

%s

This link will expire in 24 hours.

If you didn't create an account, you can safely ignore this email.

---
This is an automated message from Go-Auth.
`, firstName, link)
}

func (s *EmailService) buildPasswordResetHTML(firstName, link string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Reset Your Password</h2>
        <p>Hi %s,</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <div style="margin: 30px 0;">
            <a href="%s" style="background-color: #2196F3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #666;">%s</p>
        <p>This link will expire in 1 hour.</p>
        <p>If you didn't request a password reset, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="font-size: 12px; color: #999;">This is an automated message from Go-Auth.</p>
    </div>
</body>
</html>
`, firstName, link, link)
}

func (s *EmailService) buildPasswordResetText(firstName, link string) string {
	return fmt.Sprintf(`
Reset Your Password

Hi %s,

We received a request to reset your password. Visit this link to create a new password:

%s

This link will expire in 1 hour.

If you didn't request a password reset, you can safely ignore this email.

---
This is an automated message from Go-Auth.
`, firstName, link)
}

func (s *EmailService) buildWelcomeHTML(firstName string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Welcome to Go-Auth!</h2>
        <p>Hi %s,</p>
        <p>Welcome aboard! Your account has been successfully created.</p>
        <p>You can now use your credentials to access the platform.</p>
        <p>If you have any questions, feel free to reach out to our support team.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="font-size: 12px; color: #999;">This is an automated message from Go-Auth.</p>
    </div>
</body>
</html>
`, firstName)
}

func (s *EmailService) buildWelcomeText(firstName string) string {
	return fmt.Sprintf(`
Welcome to Go-Auth!

Hi %s,

Welcome aboard! Your account has been successfully created.

You can now use your credentials to access the platform.

If you have any questions, feel free to reach out to our support team.

---
This is an automated message from Go-Auth.
`, firstName)
}

// GenerateVerificationToken generates a secure verification token with expiry
func (s *EmailService) GenerateVerificationToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := s.client.EmailVerifications.Create().
		SetUserID(userID).
		SetEmail(email).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create verification token: %w", err)
	}

	return token, nil
}

// GeneratePasswordResetToken generates a secure password reset token with expiry
func (s *EmailService) GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := s.client.PasswordResets.Create().
		SetUserID(userID).
		SetEmail(email).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create password reset token: %w", err)
	}

	return token, nil
}
