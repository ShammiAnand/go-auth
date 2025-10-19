package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/permissions"
	"github.com/shammianand/go-auth/ent/rolepermissions"
	"github.com/shammianand/go-auth/ent/roles"
)

// BootstrapService handles RBAC initialization
type BootstrapService struct {
	client *ent.Client
	logger *slog.Logger
}

// NewBootstrapService creates a new bootstrap service
func NewBootstrapService(client *ent.Client, logger *slog.Logger) *BootstrapService {
	if logger == nil {
		logger = slog.Default()
	}

	return &BootstrapService{
		client: client,
		logger: logger,
	}
}

// BootstrapPermissions creates or updates permissions from config
func (s *BootstrapService) BootstrapPermissions(ctx context.Context, perms []PermissionConfig) (int, int, error) {
	created := 0
	updated := 0

	for _, perm := range perms {
		// Check if permission exists
		existing, err := s.client.Permissions.Query().
			Where(permissions.CodeEQ(perm.Code)).
			Only(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return created, updated, fmt.Errorf("failed to query permission %s: %w", perm.Code, err)
		}

		if existing == nil {
			// Create new permission
			_, err := s.client.Permissions.Create().
				SetCode(perm.Code).
				SetName(perm.Name).
				SetNillableDescription(&perm.Description).
				SetNillableResource(&perm.Resource).
				SetNillableAction(&perm.Action).
				Save(ctx)

			if err != nil {
				return created, updated, fmt.Errorf("failed to create permission %s: %w", perm.Code, err)
			}

			s.logger.Info("Permission created", "code", perm.Code)
			created++
		} else {
			// Update existing permission
			_, err := existing.Update().
				SetName(perm.Name).
				SetNillableDescription(&perm.Description).
				SetNillableResource(&perm.Resource).
				SetNillableAction(&perm.Action).
				Save(ctx)

			if err != nil {
				return created, updated, fmt.Errorf("failed to update permission %s: %w", perm.Code, err)
			}

			s.logger.Info("Permission updated", "code", perm.Code)
			updated++
		}
	}

	return created, updated, nil
}

// BootstrapRoles creates or updates roles from config and assigns permissions
func (s *BootstrapService) BootstrapRoles(ctx context.Context, roleConfigs []RoleConfig) (int, int, error) {
	created := 0
	updated := 0

	// Get all permissions for wildcard matching
	allPermissions, err := s.client.Permissions.Query().All(ctx)
	if err != nil {
		return created, updated, fmt.Errorf("failed to query permissions: %w", err)
	}

	for _, roleConfig := range roleConfigs {
		// Check if role exists
		existing, err := s.client.Roles.Query().
			Where(roles.CodeEQ(roleConfig.Code)).
			Only(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return created, updated, fmt.Errorf("failed to query role %s: %w", roleConfig.Code, err)
		}

		var role *ent.Roles
		if existing == nil {
			// Create new role
			role, err = s.client.Roles.Create().
				SetCode(roleConfig.Code).
				SetName(roleConfig.Name).
				SetNillableDescription(&roleConfig.Description).
				SetIsSystem(roleConfig.IsSystem).
				SetIsDefault(roleConfig.IsDefault).
				SetNillableMaxUsers(roleConfig.MaxUsers).
				Save(ctx)

			if err != nil {
				return created, updated, fmt.Errorf("failed to create role %s: %w", roleConfig.Code, err)
			}

			s.logger.Info("Role created", "code", roleConfig.Code)
			created++
		} else {
			// Update existing role
			role, err = existing.Update().
				SetName(roleConfig.Name).
				SetNillableDescription(&roleConfig.Description).
				SetIsSystem(roleConfig.IsSystem).
				SetIsDefault(roleConfig.IsDefault).
				SetNillableMaxUsers(roleConfig.MaxUsers).
				Save(ctx)

			if err != nil {
				return created, updated, fmt.Errorf("failed to update role %s: %w", roleConfig.Code, err)
			}

			s.logger.Info("Role updated", "code", roleConfig.Code)
			updated++
		}

		// Assign permissions
		err = s.assignPermissionsToRole(ctx, role, roleConfig.Permissions, allPermissions)
		if err != nil {
			return created, updated, fmt.Errorf("failed to assign permissions to role %s: %w", roleConfig.Code, err)
		}
	}

	return created, updated, nil
}

// assignPermissionsToRole assigns permissions to a role based on permission codes/wildcards
func (s *BootstrapService) assignPermissionsToRole(ctx context.Context, role *ent.Roles, permCodes []string, allPermissions []*ent.Permissions) error {
	// Resolve permission IDs from codes and wildcards
	permissionIDs := make([]int, 0)

	for _, permCode := range permCodes {
		if permCode == "*" {
			// All permissions
			for _, perm := range allPermissions {
				permissionIDs = append(permissionIDs, perm.ID)
			}
			break
		} else if strings.HasSuffix(permCode, ".*") {
			// Wildcard match (e.g., "users.*")
			prefix := strings.TrimSuffix(permCode, ".*")
			for _, perm := range allPermissions {
				if strings.HasPrefix(perm.Code, prefix+".") || perm.Code == prefix {
					permissionIDs = append(permissionIDs, perm.ID)
				}
			}
		} else {
			// Exact match
			for _, perm := range allPermissions {
				if perm.Code == permCode {
					permissionIDs = append(permissionIDs, perm.ID)
					break
				}
			}
		}
	}

	// Remove duplicates
	permissionIDs = uniqueInts(permissionIDs)

	// Get existing role-permission assignments
	existingAssignments, err := s.client.RolePermissions.Query().
		Where(rolepermissions.RoleIDEQ(role.ID)).
		All(ctx)

	if err != nil {
		return fmt.Errorf("failed to query existing role permissions: %w", err)
	}

	existingPermIDs := make(map[int]bool)
	for _, assignment := range existingAssignments {
		existingPermIDs[assignment.PermissionID] = true
	}

	// Add new permissions
	for _, permID := range permissionIDs {
		if !existingPermIDs[permID] {
			_, err := s.client.RolePermissions.Create().
				SetRoleID(role.ID).
				SetPermissionID(permID).
				Save(ctx)

			if err != nil {
				s.logger.Error("Failed to assign permission to role",
					"role_id", role.ID,
					"permission_id", permID,
					"error", err,
				)
				// Continue with other permissions
			}
		}
	}

	// Remove permissions not in config
	targetPermIDs := make(map[int]bool)
	for _, permID := range permissionIDs {
		targetPermIDs[permID] = true
	}

	for _, assignment := range existingAssignments {
		if !targetPermIDs[assignment.PermissionID] {
			err := s.client.RolePermissions.DeleteOne(assignment).Exec(ctx)
			if err != nil {
				s.logger.Error("Failed to remove permission from role",
					"role_id", role.ID,
					"permission_id", assignment.PermissionID,
					"error", err,
				)
			}
		}
	}

	return nil
}

// Helper function to remove duplicate ints
func uniqueInts(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, val := range slice {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}

	return result
}
