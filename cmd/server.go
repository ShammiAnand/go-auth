package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/common/middleware"
	"github.com/shammianand/go-auth/internal/config"
	authmodule "github.com/shammianand/go-auth/internal/modules/auth"
	"github.com/shammianand/go-auth/internal/modules/email/provider"
	emailservice "github.com/shammianand/go-auth/internal/modules/email/service"
	rbacmodule "github.com/shammianand/go-auth/internal/modules/rbac"
	"github.com/shammianand/go-auth/internal/storage"
	"github.com/spf13/cobra"
)

var (
	serverPort string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server",
	Long:  `Starts the authentication HTTP API server with Gin framework`,
	RunE:  runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "", "Port to run server on (default from config)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Connect to database
	entClient, err := storage.DBConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer entClient.Close()

	// Run auto-migration
	err = storage.AutoMigrate(*entClient)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Connect to Redis
	redisClient := storage.GetRedisClient()

	// Initialize JWKS keys
	err = auth.InitializeKeys(redisClient)
	if err != nil {
		return fmt.Errorf("failed to initialize JWKS keys: %w", err)
	}

	// Determine port
	port := serverPort
	if port == "" {
		port = config.ENV_API_PORT
	}

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check database connectivity
		if err := entClient.Schema.Create(context.Background()); err != nil {
			// Schema exists is fine
		}
		c.JSON(200, gin.H{"status": "ready", "service": "go-auth"})
	})

	// Initialize email service
	emailProvider := provider.NewMailhogProvider(
		"localhost", // TODO: from config
		"1025",      // TODO: from config
		"noreply@go-auth.local",
		logger,
	)
	emailSvc := emailservice.NewEmailService(
		emailProvider,
		entClient,
		logger,
		"noreply@go-auth.local",
		"Go-Auth",
	)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public JWKS endpoint
		v1.GET("/.well-known/jwks.json", func(c *gin.Context) {
			jwksJSON, err := redisClient.Get(context.Background(), "auth:jwks").Result()
			if err != nil {
				c.JSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
			c.Header("Content-Type", "application/json")
			c.String(200, jwksJSON)
		})

		// Register module routes
		authmodule.RegisterRoutes(v1, entClient, redisClient, emailSvc, logger)
		rbacmodule.RegisterRoutes(v1, entClient, redisClient, logger)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited gracefully")
	return nil
}
