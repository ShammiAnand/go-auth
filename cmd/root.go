package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-auth",
	Short: "Go-Auth - Lightweight authentication microservice with JWKS support",
	Long: `Go-Auth is a production-ready authentication microservice that provides:
  - JWT-based authentication with JWKS (JSON Web Key Set)
  - Role-Based Access Control (RBAC)
  - Email verification and password reset flows
  - Multi-backend support through JWKS

Perfect for microservice architectures requiring centralized authentication.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-auth.yaml)")
}
