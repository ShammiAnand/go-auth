package controller

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/shammianand/go-auth/internal/common/middleware"
	"github.com/shammianand/go-auth/internal/common/types"
	"github.com/shammianand/go-auth/internal/common/utils"
	"github.com/shammianand/go-auth/internal/modules/auth/models"
	"github.com/shammianand/go-auth/internal/modules/auth/service"
)

// AuthController handles authentication HTTP requests
type AuthController struct {
	service *service.AuthService
	logger  *slog.Logger
}

// NewAuthController creates a new auth controller
func NewAuthController(service *service.AuthService, logger *slog.Logger) *AuthController {
	return &AuthController{
		service: service,
		logger:  logger,
	}
}

// Signup handles user registration
func (ac *AuthController) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	resp, err := ac.service.Signup(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.BadRequest, "Signup failed", "SIGNUP_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Created, "User created successfully. Please check your email to verify your account.", resp)
}

// Signin handles user authentication
func (ac *AuthController) Signin(c *gin.Context) {
	var req models.SigninRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	resp, err := ac.service.Signin(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.Unauthorized, "Authentication failed", "AUTH_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Authentication successful", resp)
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.RespondError(c, types.HTTP.Unauthorized, "Not authenticated", "UNAUTHORIZED", err.Error())
		return
	}

	err = ac.service.Logout(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, types.HTTP.InternalServerError, "Logout failed", "LOGOUT_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Logged out successfully", nil)
}

// GetMe returns current user info
func (ac *AuthController) GetMe(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.RespondError(c, types.HTTP.Unauthorized, "Not authenticated", "UNAUTHORIZED", err.Error())
		return
	}

	userInfo, err := ac.service.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, types.HTTP.NotFound, "User not found", "USER_NOT_FOUND", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "User info retrieved", userInfo)
}

// UpdateProfile updates user profile
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.RespondError(c, types.HTTP.Unauthorized, "Not authenticated", "UNAUTHORIZED", err.Error())
		return
	}

	var req models.UpdateProfileRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	userInfo, err := ac.service.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.BadRequest, "Profile update failed", "UPDATE_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Profile updated successfully", userInfo)
}

// ForgotPassword initiates password reset
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	err := ac.service.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.InternalServerError, "Failed to process request", "FORGOT_PASSWORD_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword completes password reset
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	err := ac.service.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.BadRequest, "Password reset failed", "RESET_PASSWORD_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Password reset successfully", nil)
}

// VerifyEmail verifies user email
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.RespondError(c, types.HTTP.BadRequest, "Token is required", "MISSING_TOKEN", "Verification token must be provided")
		return
	}

	err := ac.service.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		utils.RespondError(c, types.HTTP.BadRequest, "Email verification failed", "VERIFICATION_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Email verified successfully", nil)
}

// ResendVerification resends verification email
func (ac *AuthController) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := utils.BindJSON(c, &req); err != nil {
		return
	}

	err := ac.service.ResendVerification(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, types.HTTP.BadRequest, "Failed to resend verification", "RESEND_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(c, types.HTTP.Ok, "Verification email sent", nil)
}
