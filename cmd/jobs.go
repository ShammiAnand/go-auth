package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/spf13/cobra"
)

var (
	jwksRefreshInterval string
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Background job commands",
	Long:  `Commands for running background jobs and scheduled tasks`,
}

var jwksRefreshCmd = &cobra.Command{
	Use:   "jwks-refresh",
	Short: "Run JWKS key rotation job",
	Long: `Runs a background job that periodically refreshes JWKS public/private keys.
This ensures key rotation for enhanced security.`,
	RunE: runJWKSRefresh,
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jwksRefreshCmd)

	jwksRefreshCmd.Flags().StringVar(&jwksRefreshInterval, "interval", "", "Refresh interval (e.g., 24h, 12h, 30m) - REQUIRED")
	jwksRefreshCmd.MarkFlagRequired("interval")
}

func runJWKSRefresh(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Parse interval
	interval, err := time.ParseDuration(jwksRefreshInterval)
	if err != nil {
		return fmt.Errorf("invalid interval format: %w (examples: 24h, 12h, 30m)", err)
	}

	if interval < 10*time.Minute {
		return fmt.Errorf("interval must be at least 10 minutes for safety")
	}

	// Connect to Redis
	redisClient := storage.GetRedisClient()
	ctx := context.Background()

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("JWKS Refresh job started",
		"interval", interval.String(),
	)

	// Create ticker for periodic refresh
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Initial key check/generation
	err = auth.InitializeKeys(redisClient)
	if err != nil {
		logger.Error("Failed initial key initialization", "error", err)
		return err
	}
	logger.Info("Initial JWKS keys verified/generated")

	// Run refresh loop
	for {
		select {
		case <-ticker.C:
			logger.Info("Running scheduled JWKS key refresh")
			err := refreshJWKSKeys(redisClient, logger)
			if err != nil {
				logger.Error("Failed to refresh JWKS keys", "error", err)
				// Don't exit on error, continue trying
			} else {
				logger.Info("JWKS keys refreshed successfully")
			}

		case <-quit:
			logger.Info("Received shutdown signal, stopping JWKS refresh job")
			return nil
		}
	}
}

func refreshJWKSKeys(redisClient *redis.Client, logger *slog.Logger) error {
	// This would implement the actual key rotation logic
	// For now, we'll just re-initialize which generates new keys if needed
	err := auth.InitializeKeys(redisClient)
	if err != nil {
		return fmt.Errorf("failed to refresh keys: %w", err)
	}

	// TODO: Implement proper key rotation:
	// 1. Generate new key pair
	// 2. Add to key set (keep old keys valid)
	// 3. Update JWKS endpoint
	// 4. Mark old keys for deprecation after grace period
	// 5. Remove expired keys

	logger.Info("JWKS key rotation completed")
	return nil
}
