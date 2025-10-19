package models

import (
	"time"

	"github.com/google/uuid"
)

// RoleResponse represents a role
type RoleResponse struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	IsDefault   bool      `json:"is_default"`
	MaxUsers    *int      `json:"max_users,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleWithPermissionsResponse includes permissions
type RoleWithPermissionsResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}

// PermissionResponse represents a permission
type PermissionResponse struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Resource    string    `json:"resource,omitempty"`
	Action      string    `json:"action,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRolesResponse represents user's roles
type UserRolesResponse struct {
	UserID      uuid.UUID      `json:"user_id"`
	Email       string         `json:"email"`
	Roles       []RoleResponse `json:"roles"`
	AssignedAt  time.Time      `json:"assigned_at"`
}

// UserPermissionsResponse represents computed user permissions
type UserPermissionsResponse struct {
	UserID      uuid.UUID            `json:"user_id"`
	Permissions []PermissionResponse `json:"permissions"`
}

// AuditLogResponse represents an audit log entry
type AuditLogResponse struct {
	ID           uuid.UUID              `json:"id"`
	ActorID      *uuid.UUID             `json:"actor_id,omitempty"`
	ActionType   string                 `json:"action_type"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}
