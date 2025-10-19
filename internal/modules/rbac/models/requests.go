package models

import "github.com/google/uuid"

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID int       `json:"role_id" binding:"required"`
}

// RemoveRoleRequest represents a request to remove a role from a user
type RemoveRoleRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RoleID int       `json:"role_id" binding:"required"`
}

// UpdateRolePermissionsRequest updates permissions for a role
type UpdateRolePermissionsRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	ActorID      string `form:"actor_id"`
	ActionType   string `form:"action_type"`
	ResourceType string `form:"resource_type"`
	ResourceID   string `form:"resource_id"`
	Limit        int    `form:"limit"`
	Offset       int    `form:"offset"`
}
