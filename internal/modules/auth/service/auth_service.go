package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/emailverifications"
	"github.com/shammianand/go-auth/ent/passwordresets"
	"github.com/shammianand/go-auth/ent/roles"
	"github.com/shammianand/go-auth/ent/users"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/modules/auth/models"
	"github.com/shammianand/go-auth/internal/modules/email/service"
)

// AuthService handles authentication operations
type AuthService struct {
	client       *ent.Client
	cache        *redis.Client
	emailService *service.EmailService
	logger       *slog.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(client *ent.Client, cache *redis.Client, emailService *service.EmailService, logger *slog.Logger) *AuthService {
	if logger == nil {
		logger = slog.Default()
	}

	return &AuthService{
		client:       client,
		cache:        cache,
		emailService: emailService,
		logger:       logger,
	}
}

// Signup creates a new user account
func (s *AuthService) Signup(ctx context.Context, req *models.SignupRequest) (*models.SignupResponse, error) {
	// Check if user already exists
	exists, err := s.client.Users.Query().
		Where(users.EmailEQ(req.Email)).
		Exist(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := auth.HashPasswords(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.client.Users.Create().
		SetEmail(req.Email).
		SetPasswordHash(hashedPassword).
		SetFirstName(req.FirstName).
		SetLastName(req.LastName).
		SetIsActive(true).
		SetEmailVerified(false).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default role
	defaultRole, err := s.client.Roles.Query().
		Where(roles.IsDefaultEQ(true)).
		First(ctx)

	if err != nil {
		s.logger.Warn("No default role found, user created without role", "user_id", user.ID)
	} else {
		_, err = s.client.UserRoles.Create().
			SetUserID(user.ID).
			SetRoleID(defaultRole.ID).
			Save(ctx)

		if err != nil {
			s.logger.Error("Failed to assign default role", "user_id", user.ID, "error", err)
		}
	}

	// Generate verification token
	token, err := s.emailService.GenerateVerificationToken(ctx, user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate verification token", "user_id", user.ID, "error", err)
	} else {
		// Send verification email
		err = s.emailService.SendVerificationEmail(ctx, user.ID, user.Email, user.FirstName, token)
		if err != nil {
			s.logger.Error("Failed to send verification email", "user_id", user.ID, "error", err)
		}
	}

	return &models.SignupResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// Signin authenticates a user and returns a JWT token
func (s *AuthService) Signin(ctx context.Context, req *models.SigninRequest) (*models.SigninResponse, error) {
	// Find user by email
	user, err := s.client.Users.Query().
		Where(users.EmailEQ(req.Email)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if !auth.ComparePasswords(user.PasswordHash, []byte(req.Password)) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	user, err = user.Update().
		SetLastLogin(time.Now()).
		Save(ctx)

	if err != nil {
		s.logger.Error("Failed to update last login", "user_id", user.ID, "error", err)
	}

	// Generate JWT
	token, err := auth.CreateJWT(user.ID, s.cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Token expires in configured time
	expiresAt := time.Now().Add(30 * time.Minute) // TODO: Get from config

	return &models.SigninResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: models.UserInfo{
			ID:            user.ID,
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			EmailVerified: user.EmailVerified,
			IsActive:      user.IsActive,
			CreatedAt:     user.CreatedAt,
			LastLogin:     user.LastLogin,
		},
	}, nil
}

// Logout invalidates a user's token
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	// Remove token from Redis
	key := fmt.Sprintf("auth:token:%s", userID.String())
	err := s.cache.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	return nil
}

// GetUserInfo retrieves user information
func (s *AuthService) GetUserInfo(ctx context.Context, userID uuid.UUID) (*models.UserInfo, error) {
	user, err := s.client.Users.Query().
		Where(users.IDEQ(userID)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &models.UserInfo{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		LastLogin:     user.LastLogin,
	}, nil
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.UserInfo, error) {
	user, err := s.client.Users.Query().
		Where(users.IDEQ(userID)).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	update := user.Update()

	if req.FirstName != nil {
		update = update.SetFirstName(*req.FirstName)
	}

	if req.LastName != nil {
		update = update.SetLastName(*req.LastName)
	}

	if req.Password != nil {
		hashedPassword, err := auth.HashPasswords(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		update = update.SetPasswordHash(hashedPassword)
	}

	user, err = update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUserInfo(ctx, userID)
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error {
	// Find user by email
	user, err := s.client.Users.Query().
		Where(users.EmailEQ(req.Email)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Don't reveal if user exists
			s.logger.Info("Password reset requested for non-existent email", "email", req.Email)
			return nil
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Generate reset token
	token, err := s.emailService.GeneratePasswordResetToken(ctx, user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Send reset email
	err = s.emailService.SendPasswordResetEmail(ctx, user.ID, user.Email, user.FirstName, token)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

// ResetPassword completes password reset
func (s *AuthService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	// Find valid reset token
	resetRecord, err := s.client.PasswordResets.Query().
		Where(
			passwordresets.TokenEQ(req.Token),
			passwordresets.IsUsedEQ(false),
			passwordresets.ExpiresAtGT(time.Now()),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("failed to find reset token: %w", err)
	}

	// Hash new password
	hashedPassword, err := auth.HashPasswords(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	_, err = s.client.Users.Update().
		Where(users.IDEQ(resetRecord.UserID)).
		SetPasswordHash(hashedPassword).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	now := time.Now()
	_, err = resetRecord.Update().
		SetIsUsed(true).
		SetUsedAt(now).
		Save(ctx)

	if err != nil {
		s.logger.Error("Failed to mark reset token as used", "error", err)
	}

	return nil
}

// VerifyEmail verifies a user's email address
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// Find valid verification token
	verifyRecord, err := s.client.EmailVerifications.Query().
		Where(
			emailverifications.TokenEQ(token),
			emailverifications.IsUsedEQ(false),
			emailverifications.ExpiresAtGT(time.Now()),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("invalid or expired verification token")
		}
		return fmt.Errorf("failed to find verification token: %w", err)
	}

	// Update user email_verified status
	_, err = s.client.Users.Update().
		Where(users.IDEQ(verifyRecord.UserID)).
		SetEmailVerified(true).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Mark token as used
	now := time.Now()
	_, err = verifyRecord.Update().
		SetIsUsed(true).
		SetUsedAt(now).
		Save(ctx)

	if err != nil {
		s.logger.Error("Failed to mark verification token as used", "error", err)
	}

	// Send welcome email
	user, _ := s.client.Users.Get(ctx, verifyRecord.UserID)
	if user != nil {
		_ = s.emailService.SendWelcomeEmail(ctx, user.ID, user.Email, user.FirstName)
	}

	return nil
}

// ResendVerification resends email verification
func (s *AuthService) ResendVerification(ctx context.Context, req *models.ResendVerificationRequest) error {
	// Find user
	user, err := s.client.Users.Query().
		Where(users.EmailEQ(req.Email)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Don't reveal if user exists
			return nil
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Check if already verified
	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate new verification token
	token, err := s.emailService.GenerateVerificationToken(ctx, user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(ctx, user.ID, user.Email, user.FirstName, token)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
