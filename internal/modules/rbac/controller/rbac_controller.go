package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shammianand/go-auth/internal/common/middleware"
	"github.com/shammianand/go-auth/internal/common/types"
	"github.com/shammianand/go-auth/internal/common/utils"
	"github.com/shammianand/go-auth/internal/modules/rbac/models"
	"github.com/shammianand/go-auth/internal/modules/rbac/service"
)

// RBACController handles RBAC HTTP requests
type RBACController struct {
	service *service.RBACService
}

// NewRBACController creates a new RBAC controller
func NewRBACController(service *service.RBACService) *RBACController {
	return &RBACController{
		service: service,
	}
}

// ListRoles returns all roles
func (c *RBACController) ListRoles(ctx *gin.Context) {
	roles, err := c.service.ListRoles(ctx.Request.Context())
	if err != nil {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to list roles", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Roles retrieved successfully", roles)
}

// GetRole returns a specific role with permissions
func (c *RBACController) GetRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.BadRequest, "Invalid role ID", "VALIDATION_ERROR", err.Error())
		return
	}

	role, err := c.service.GetRole(ctx.Request.Context(), roleID)
	if err != nil {
		if err.Error() == "role not found" {
			utils.RespondError(ctx, types.HTTP.NotFound, "Role not found", "ROLE_NOT_FOUND", err.Error())
			return
		}
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to get role", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Role retrieved successfully", role)
}

// ListPermissions returns all permissions
func (c *RBACController) ListPermissions(ctx *gin.Context) {
	permissions, err := c.service.ListPermissions(ctx.Request.Context())
	if err != nil {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to list permissions", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Permissions retrieved successfully", permissions)
}

// GetUserRoles returns roles assigned to a user
func (c *RBACController) GetUserRoles(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.BadRequest, "Invalid user ID", "VALIDATION_ERROR", err.Error())
		return
	}

	userRoles, err := c.service.GetUserRoles(ctx.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.RespondError(ctx, types.HTTP.NotFound, "User not found", "USER_NOT_FOUND", err.Error())
			return
		}
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to get user roles", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "User roles retrieved successfully", userRoles)
}

// AssignRole assigns a role to a user
func (c *RBACController) AssignRole(ctx *gin.Context) {
	var req models.AssignRoleRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return
	}

	// Get actor ID from context (set by auth middleware)
	actorID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		utils.RespondError(ctx, types.HTTP.Unauthorized, "Authentication required", "UNAUTHORIZED", "Actor ID not found")
		return
	}

	actorUUID, ok := actorID.(uuid.UUID)
	if !ok {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Invalid actor ID format", "INTERNAL_ERROR", "Actor ID type mismatch")
		return
	}

	err := c.service.AssignRole(ctx.Request.Context(), req.UserID, req.RoleID, actorUUID)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "role not found" {
			utils.RespondError(ctx, types.HTTP.NotFound, err.Error(), "NOT_FOUND", err.Error())
			return
		}
		if err.Error() == "role already assigned to user" {
			utils.RespondError(ctx, types.HTTP.Conflict, err.Error(), "CONFLICT", err.Error())
			return
		}
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to assign role", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Role assigned successfully", nil)
}

// RemoveRole removes a role from a user
func (c *RBACController) RemoveRole(ctx *gin.Context) {
	var req models.RemoveRoleRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return
	}

	// Get actor ID from context
	actorID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		utils.RespondError(ctx, types.HTTP.Unauthorized, "Authentication required", "UNAUTHORIZED", "Actor ID not found")
		return
	}

	actorUUID, ok := actorID.(uuid.UUID)
	if !ok {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Invalid actor ID format", "INTERNAL_ERROR", "Actor ID type mismatch")
		return
	}

	err := c.service.RemoveRole(ctx.Request.Context(), req.UserID, req.RoleID, actorUUID)
	if err != nil {
		if err.Error() == "role assignment not found" {
			utils.RespondError(ctx, types.HTTP.NotFound, err.Error(), "NOT_FOUND", err.Error())
			return
		}
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to remove role", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Role removed successfully", nil)
}

// GetUserPermissions returns computed permissions for a user
func (c *RBACController) GetUserPermissions(ctx *gin.Context) {
	userIDStr := ctx.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.BadRequest, "Invalid user ID", "VALIDATION_ERROR", err.Error())
		return
	}

	permissions, err := c.service.GetUserPermissions(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to get user permissions", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "User permissions retrieved successfully", permissions)
}

// UpdateRolePermissions updates permissions for a role
func (c *RBACController) UpdateRolePermissions(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.BadRequest, "Invalid role ID", "VALIDATION_ERROR", err.Error())
		return
	}

	var req models.UpdateRolePermissionsRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return
	}

	// Get actor ID from context
	actorID, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		utils.RespondError(ctx, types.HTTP.Unauthorized, "Authentication required", "UNAUTHORIZED", "Actor ID not found")
		return
	}

	actorUUID, ok := actorID.(uuid.UUID)
	if !ok {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Invalid actor ID format", "INTERNAL_ERROR", "Actor ID type mismatch")
		return
	}

	err = c.service.UpdateRolePermissions(ctx.Request.Context(), roleID, req.PermissionIDs, actorUUID)
	if err != nil {
		if err.Error() == "role not found" {
			utils.RespondError(ctx, types.HTTP.NotFound, err.Error(), "NOT_FOUND", err.Error())
			return
		}
		if err.Error() == "cannot modify permissions of system role" {
			utils.RespondError(ctx, types.HTTP.Forbidden, err.Error(), "FORBIDDEN", err.Error())
			return
		}
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to update role permissions", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Role permissions updated successfully", nil)
}

// GetAuditLogs returns audit logs with filters
func (c *RBACController) GetAuditLogs(ctx *gin.Context) {
	var filter models.AuditLogFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		utils.RespondError(ctx, types.HTTP.BadRequest, "Invalid query parameters", "VALIDATION_ERROR", err.Error())
		return
	}

	logs, err := c.service.GetAuditLogs(ctx.Request.Context(), &filter)
	if err != nil {
		utils.RespondError(ctx, types.HTTP.InternalServerError, "Failed to get audit logs", "RBAC_ERROR", err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.HTTP.Ok, "Audit logs retrieved successfully", logs)
}
