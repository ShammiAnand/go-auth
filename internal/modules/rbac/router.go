package rbac

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/internal/common/middleware"
	"github.com/shammianand/go-auth/internal/modules/rbac/controller"
	"github.com/shammianand/go-auth/internal/modules/rbac/service"
)

// RegisterRoutes registers RBAC routes
func RegisterRoutes(
	router *gin.RouterGroup,
	client *ent.Client,
	redisClient *redis.Client,
	logger *slog.Logger,
) {
	// Initialize service and controller
	rbacService := service.NewRBACService(client, logger)
	rbacController := controller.NewRBACController(rbacService)

	// Create rbac group under /api/v1/rbac
	rbac := router.Group("/rbac")

	// Public routes (read-only, no auth required)
	rbac.GET("/roles", rbacController.ListRoles)
	rbac.GET("/roles/:id", rbacController.GetRole)
	rbac.GET("/permissions", rbacController.ListPermissions)

	// Protected routes (require authentication)
	authenticated := rbac.Group("")
	authenticated.Use(middleware.RequireAuth(redisClient))
	{
		// User roles and permissions (any authenticated user can view)
		authenticated.GET("/users/:user_id/roles", rbacController.GetUserRoles)
		authenticated.GET("/users/:user_id/permissions", rbacController.GetUserPermissions)

		// Role assignment (require admin permissions)
		authenticated.POST("/users/assign-role", rbacController.AssignRole)
		authenticated.POST("/users/remove-role", rbacController.RemoveRole)

		// Role permission management (require admin permissions)
		authenticated.PUT("/roles/:id/permissions", rbacController.UpdateRolePermissions)

		// Audit logs (require admin permissions)
		authenticated.GET("/audit-logs", rbacController.GetAuditLogs)
	}
}
