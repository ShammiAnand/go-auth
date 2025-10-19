package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/auditlogs"
	"github.com/shammianand/go-auth/ent/rolepermissions"
	"github.com/shammianand/go-auth/ent/userroles"
	"github.com/shammianand/go-auth/ent/users"
	"github.com/shammianand/go-auth/internal/modules/rbac/models"
)

// RBACService handles RBAC operations
type RBACService struct {
	client *ent.Client
	logger *slog.Logger
}

// NewRBACService creates a new RBAC service
func NewRBACService(client *ent.Client, logger *slog.Logger) *RBACService {
	if logger == nil {
		logger = slog.Default()
	}

	return &RBACService{
		client: client,
		logger: logger,
	}
}

// ListRoles returns all roles
func (s *RBACService) ListRoles(ctx context.Context) ([]models.RoleResponse, error) {
	entRoles, err := s.client.Roles.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	result := make([]models.RoleResponse, len(entRoles))
	for i, role := range entRoles {
		result[i] = s.roleToResponse(role)
	}

	return result, nil
}

// GetRole returns a role with its permissions
func (s *RBACService) GetRole(ctx context.Context, roleID int) (*models.RoleWithPermissionsResponse, error) {
	role, err := s.client.Roles.Get(ctx, roleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Get role permissions
	rolePerms, err := s.client.RolePermissions.Query().
		Where(rolepermissions.RoleIDEQ(roleID)).
		WithPermission().
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	perms := make([]models.PermissionResponse, len(rolePerms))
	for i, rp := range rolePerms {
		if rp.Edges.Permission != nil {
			perms[i] = s.permissionToResponse(rp.Edges.Permission)
		}
	}

	return &models.RoleWithPermissionsResponse{
		RoleResponse: s.roleToResponse(role),
		Permissions:  perms,
	}, nil
}

// ListPermissions returns all permissions
func (s *RBACService) ListPermissions(ctx context.Context) ([]models.PermissionResponse, error) {
	entPerms, err := s.client.Permissions.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	result := make([]models.PermissionResponse, len(entPerms))
	for i, perm := range entPerms {
		result[i] = s.permissionToResponse(perm)
	}

	return result, nil
}

// GetUserRoles returns roles assigned to a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) (*models.UserRolesResponse, error) {
	user, err := s.client.Users.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	userRolesList, err := s.client.UserRoles.Query().
		Where(userroles.UserIDEQ(userID)).
		WithRole().
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleResponses := make([]models.RoleResponse, 0, len(userRolesList))
	var latestAssignment time.Time

	for _, ur := range userRolesList {
		if ur.Edges.Role != nil {
			roleResponses = append(roleResponses, s.roleToResponse(ur.Edges.Role))
			if ur.AssignedAt.After(latestAssignment) {
				latestAssignment = ur.AssignedAt
			}
		}
	}

	return &models.UserRolesResponse{
		UserID:     userID,
		Email:      user.Email,
		Roles:      roleResponses,
		AssignedAt: latestAssignment,
	}, nil
}

// AssignRole assigns a role to a user
func (s *RBACService) AssignRole(ctx context.Context, userID uuid.UUID, roleID int, actorID uuid.UUID) error {
	// Check if user exists
	exists, err := s.client.Users.Query().Where(users.IDEQ(userID)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Check if role exists
	role, err := s.client.Roles.Get(ctx, roleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Check max_users constraint
	if role.MaxUsers != nil {
		count, err := s.client.UserRoles.Query().
			Where(userroles.RoleIDEQ(roleID)).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count role users: %w", err)
		}
		if count >= *role.MaxUsers {
			return fmt.Errorf("role has reached maximum users limit (%d)", *role.MaxUsers)
		}
	}

	// Check if already assigned
	exists, err = s.client.UserRoles.Query().
		Where(
			userroles.UserIDEQ(userID),
			userroles.RoleIDEQ(roleID),
		).
		Exist(ctx)

	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}

	if exists {
		return fmt.Errorf("role already assigned to user")
	}

	// Create assignment
	_, err = s.client.UserRoles.Create().
		SetUserID(userID).
		SetRoleID(roleID).
		SetNillableAssignedBy(&actorID).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, actorID, "role.assign", "user_role", userID.String(), map[string]interface{}{
		"user_id": userID.String(),
		"role_id": roleID,
	})

	return nil
}

// RemoveRole removes a role from a user
func (s *RBACService) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int, actorID uuid.UUID) error {
	// Find the assignment
	assignment, err := s.client.UserRoles.Query().
		Where(
			userroles.UserIDEQ(userID),
			userroles.RoleIDEQ(roleID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role assignment not found")
		}
		return fmt.Errorf("failed to find assignment: %w", err)
	}

	// Delete assignment
	err = s.client.UserRoles.DeleteOne(assignment).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, actorID, "role.remove", "user_role", userID.String(), map[string]interface{}{
		"user_id": userID.String(),
		"role_id": roleID,
	})

	return nil
}

// GetUserPermissions returns computed permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) (*models.UserPermissionsResponse, error) {
	// Get all user roles
	userRolesList, err := s.client.UserRoles.Query().
		Where(userroles.UserIDEQ(userID)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(userRolesList) == 0 {
		return &models.UserPermissionsResponse{
			UserID:      userID,
			Permissions: []models.PermissionResponse{},
		}, nil
	}

	// Get role IDs
	roleIDs := make([]int, len(userRolesList))
	for i, ur := range userRolesList {
		roleIDs[i] = ur.RoleID
	}

	// Get all permissions for these roles
	rolePerms, err := s.client.RolePermissions.Query().
		Where(rolepermissions.RoleIDIn(roleIDs...)).
		WithPermission().
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Deduplicate permissions
	permMap := make(map[int]*ent.Permissions)
	for _, rp := range rolePerms {
		if rp.Edges.Permission != nil {
			permMap[rp.PermissionID] = rp.Edges.Permission
		}
	}

	perms := make([]models.PermissionResponse, 0, len(permMap))
	for _, perm := range permMap {
		perms = append(perms, s.permissionToResponse(perm))
	}

	return &models.UserPermissionsResponse{
		UserID:      userID,
		Permissions: perms,
	}, nil
}

// UpdateRolePermissions updates permissions for a role
func (s *RBACService) UpdateRolePermissions(ctx context.Context, roleID int, permissionIDs []int, actorID uuid.UUID) error {
	// Check if role exists and is not system role
	role, err := s.client.Roles.Get(ctx, roleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("role not found")
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return fmt.Errorf("cannot modify permissions of system role")
	}

	// Get existing permissions
	existing, err := s.client.RolePermissions.Query().
		Where(rolepermissions.RoleIDEQ(roleID)).
		All(ctx)

	if err != nil {
		return fmt.Errorf("failed to get existing permissions: %w", err)
	}

	existingMap := make(map[int]bool)
	for _, rp := range existing {
		existingMap[rp.PermissionID] = true
	}

	targetMap := make(map[int]bool)
	for _, pid := range permissionIDs {
		targetMap[pid] = true
	}

	// Add new permissions
	for _, permID := range permissionIDs {
		if !existingMap[permID] {
			_, err := s.client.RolePermissions.Create().
				SetRoleID(roleID).
				SetPermissionID(permID).
				Save(ctx)
			if err != nil {
				s.logger.Error("Failed to add permission", "role_id", roleID, "permission_id", permID, "error", err)
			}
		}
	}

	// Remove old permissions
	for _, rp := range existing {
		if !targetMap[rp.PermissionID] {
			err := s.client.RolePermissions.DeleteOne(rp).Exec(ctx)
			if err != nil {
				s.logger.Error("Failed to remove permission", "role_id", roleID, "permission_id", rp.PermissionID, "error", err)
			}
		}
	}

	// Create audit log
	s.createAuditLog(ctx, actorID, "role.permissions.update", "role", fmt.Sprintf("%d", roleID), map[string]interface{}{
		"role_id":        roleID,
		"permission_ids": permissionIDs,
	})

	return nil
}

// GetAuditLogs returns audit logs with filters
func (s *RBACService) GetAuditLogs(ctx context.Context, filter *models.AuditLogFilter) ([]models.AuditLogResponse, error) {
	query := s.client.AuditLogs.Query()

	// Apply filters
	if filter.ActorID != "" {
		actorUUID, err := uuid.Parse(filter.ActorID)
		if err == nil {
			query = query.Where(auditlogs.ActorIDEQ(actorUUID))
		}
	}

	if filter.ActionType != "" {
		query = query.Where(auditlogs.ActionTypeEQ(filter.ActionType))
	}

	if filter.ResourceType != "" {
		query = query.Where(auditlogs.ResourceTypeEQ(filter.ResourceType))
	}

	if filter.ResourceID != "" {
		query = query.Where(auditlogs.ResourceIDEQ(filter.ResourceID))
	}

	// Pagination
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	logs, err := query.
		Limit(filter.Limit).
		Offset(filter.Offset).
		Order(ent.Desc(auditlogs.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	result := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		result[i] = s.auditLogToResponse(log)
	}

	return result, nil
}

// Helper functions

func (s *RBACService) roleToResponse(role *ent.Roles) models.RoleResponse {
	return models.RoleResponse{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		IsDefault:   role.IsDefault,
		MaxUsers:    role.MaxUsers,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (s *RBACService) permissionToResponse(perm *ent.Permissions) models.PermissionResponse {
	return models.PermissionResponse{
		ID:          perm.ID,
		Code:        perm.Code,
		Name:        perm.Name,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		CreatedAt:   perm.CreatedAt,
	}
}

func (s *RBACService) auditLogToResponse(log *ent.AuditLogs) models.AuditLogResponse {
	return models.AuditLogResponse{
		ID:           log.ID,
		ActorID:      log.ActorID,
		ActionType:   log.ActionType,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		Metadata:     log.Metadata,
		Changes:      log.Changes,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		CreatedAt:    log.CreatedAt,
	}
}

func (s *RBACService) createAuditLog(ctx context.Context, actorID uuid.UUID, actionType, resourceType, resourceID string, metadata map[string]interface{}) {
	_, err := s.client.AuditLogs.Create().
		SetActorID(actorID).
		SetActionType(actionType).
		SetResourceType(resourceType).
		SetNillableResourceID(&resourceID).
		SetMetadata(metadata).
		Save(ctx)

	if err != nil {
		s.logger.Error("Failed to create audit log",
			"actor_id", actorID,
			"action", actionType,
			"error", err,
		)
	}
}
