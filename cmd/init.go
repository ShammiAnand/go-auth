package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/shammianand/go-auth/internal/modules/rbac/bootstrap"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configPath string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize roles and permissions",
	Long: `Initialize the RBAC system by creating default roles and permissions from a config file.
This operation is idempotent - it can be run multiple times safely.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&configPath, "config", "c", "./configs/rbac-config.yaml", "Path to RBAC configuration file")
}

func runInit(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting RBAC initialization", "config", configPath)

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config bootstrap.RBACConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Connect to database
	entClient, err := storage.DBConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer entClient.Close()

	// Run migrations first
	err = storage.AutoMigrate(*entClient)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	ctx := context.Background()

	// Initialize RBAC bootstrap service
	bootstrapService := bootstrap.NewBootstrapService(entClient, logger)

	// Bootstrap permissions
	logger.Info("Bootstrapping permissions", "count", len(config.Permissions))
	createdPerms, updatedPerms, err := bootstrapService.BootstrapPermissions(ctx, config.Permissions)
	if err != nil {
		return fmt.Errorf("failed to bootstrap permissions: %w", err)
	}
	logger.Info("Permissions bootstrapped",
		"created", createdPerms,
		"updated", updatedPerms,
		"total", len(config.Permissions),
	)

	// Bootstrap roles
	logger.Info("Bootstrapping roles", "count", len(config.Roles))
	createdRoles, updatedRoles, err := bootstrapService.BootstrapRoles(ctx, config.Roles)
	if err != nil {
		return fmt.Errorf("failed to bootstrap roles: %w", err)
	}
	logger.Info("Roles bootstrapped",
		"created", createdRoles,
		"updated", updatedRoles,
		"total", len(config.Roles),
	)

	fmt.Printf("\n✅ RBAC initialization completed successfully!\n\n")
	fmt.Printf("   Permissions: %d created, %d updated\n", createdPerms, updatedPerms)
	fmt.Printf("   Roles: %d created, %d updated\n\n", createdRoles, updatedRoles)

	return nil
}

func validateConfig(config *bootstrap.RBACConfig) error {
	if len(config.Permissions) == 0 {
		return fmt.Errorf("no permissions defined in config")
	}

	if len(config.Roles) == 0 {
		return fmt.Errorf("no roles defined in config")
	}

	// Validate permission codes are unique
	permCodes := make(map[string]bool)
	for _, perm := range config.Permissions {
		if perm.Code == "" {
			return fmt.Errorf("permission code cannot be empty")
		}
		if permCodes[perm.Code] {
			return fmt.Errorf("duplicate permission code: %s", perm.Code)
		}
		permCodes[perm.Code] = true
	}

	// Validate role codes are unique
	roleCodes := make(map[string]bool)
	for _, role := range config.Roles {
		if role.Code == "" {
			return fmt.Errorf("role code cannot be empty")
		}
		if roleCodes[role.Code] {
			return fmt.Errorf("duplicate role code: %s", role.Code)
		}
		roleCodes[role.Code] = true
	}

	// Ensure at least one default role
	hasDefault := false
	for _, role := range config.Roles {
		if role.IsDefault {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		return fmt.Errorf("at least one role must be marked as default (is_default: true)")
	}

	return nil
}
