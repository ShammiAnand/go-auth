package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/shammianand/go-auth/ent"
	"github.com/shammianand/go-auth/ent/roles"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/spf13/cobra"
)

var (
	adminEmail     string
	adminPassword  string
	adminFirstName string
	adminLastName  string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management commands",
	Long:  `Commands for managing administrative users and operations`,
}

var createSuperuserCmd = &cobra.Command{
	Use:   "create-superuser",
	Short: "Create a superuser with full administrative privileges",
	Long: `Creates a new superuser account with the super-admin role.
Only one super-admin can exist in the system at a time.`,
	RunE: createSuperuser,
}

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(createSuperuserCmd)

	createSuperuserCmd.Flags().StringVar(&adminEmail, "email", "", "Admin email (required)")
	createSuperuserCmd.Flags().StringVar(&adminPassword, "password", "", "Admin password (required)")
	createSuperuserCmd.Flags().StringVar(&adminFirstName, "first-name", "", "Admin first name (required)")
	createSuperuserCmd.Flags().StringVar(&adminLastName, "last-name", "", "Admin last name (required)")

	createSuperuserCmd.MarkFlagRequired("email")
	createSuperuserCmd.MarkFlagRequired("password")
	createSuperuserCmd.MarkFlagRequired("first-name")
	createSuperuserCmd.MarkFlagRequired("last-name")
}

func createSuperuser(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	entClient, err := storage.DBConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer entClient.Close()

	ctx := context.Background()

	superAdminRole, err := entClient.Roles.Query().
		Where(roles.CodeEQ("super-admin")).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("super-admin role not found. Please run 'go-auth init' first to bootstrap roles and permissions")
		}
		return fmt.Errorf("failed to query super-admin role: %w", err)
	}

	// Check if there's already a super-admin
	if superAdminRole.MaxUsers != nil && *superAdminRole.MaxUsers == 1 {
		existingCount, err := entClient.UserRoles.Query().
			Where(
			// TODO: Add filter by role_id when UserRoles edges are properly set up
			).
			Count(ctx)

		if err == nil && existingCount >= 1 {
			return fmt.Errorf("super-admin user already exists. Only one super-admin is allowed")
		}
	}

	// Hash password
	hashedPassword, err := auth.HashPasswords(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := entClient.Users.Create().
		SetEmail(adminEmail).
		SetPasswordHash(hashedPassword).
		SetFirstName(adminFirstName).
		SetLastName(adminLastName).
		SetIsActive(true).
		SetEmailVerified(true). // Super admin is auto-verified
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	_, err = entClient.UserRoles.Create().
		SetUserID(user.ID).
		SetRoleID(superAdminRole.ID).
		Save(ctx)

	if err != nil {
		entClient.Users.DeleteOne(user).Exec(ctx)
		return fmt.Errorf("failed to assign super-admin role: %w", err)
	}

	logger.Info("Super-admin user created successfully",
		"email", adminEmail,
		"user_id", user.ID,
		"role", "super-admin",
	)

	fmt.Printf("\n✅ Super-admin created successfully!\n")
	fmt.Printf("   Email: %s\n", adminEmail)
	fmt.Printf("   Name: %s %s\n", adminFirstName, adminLastName)
	fmt.Printf("   User ID: %s\n\n", user.ID)

	return nil
}
