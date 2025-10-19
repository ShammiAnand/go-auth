package auth

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/internal/common/middleware"
	"github.com/shammianand/go-auth/internal/modules/auth/controller"
	"github.com/shammianand/go-auth/internal/modules/auth/service"
	emailService "github.com/shammianand/go-auth/internal/modules/email/service"
)

// RegisterRoutes registers auth module routes
func RegisterRoutes(router *gin.RouterGroup, client *ent.Client, cache *redis.Client, emailSvc *emailService.EmailService, logger *slog.Logger) {
	// Initialize auth service and controller
	authService := service.NewAuthService(client, cache, emailSvc, logger)
	authController := controller.NewAuthController(authService, logger)

	// Public routes (no authentication required)
	auth := router.Group("/auth")
	{
		auth.POST("/signup", authController.Signup)
		auth.POST("/signin", authController.Signin)
		auth.POST("/forgot-password", authController.ForgotPassword)
		auth.POST("/reset-password", authController.ResetPassword)
		auth.GET("/verify-email", authController.VerifyEmail)
		auth.POST("/resend-verification", authController.ResendVerification)
	}

	// Protected routes (authentication required)
	authProtected := router.Group("/auth")
	authProtected.Use(middleware.RequireAuth(cache))
	{
		authProtected.POST("/logout", authController.Logout)
		authProtected.GET("/me", authController.GetMe)
		authProtected.PUT("/me", authController.UpdateProfile)
	}
}
