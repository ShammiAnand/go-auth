package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shammianand/go-auth/internal/auth"
	"github.com/shammianand/go-auth/internal/common/types"
	"github.com/shammianand/go-auth/internal/common/utils"
)

const UserIDKey = "user_id"

// RequireAuth middleware validates JWT tokens and sets user_id in context
func RequireAuth(cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondError(c, types.HTTP.Unauthorized, "Authentication required", "MISSING_AUTH_HEADER", "Authorization header is required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondError(c, types.HTTP.Unauthorized, "Invalid authorization header", "INVALID_AUTH_HEADER", "Format must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get public key from cache
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing kid in token header")
			}

			return auth.GetPublicKeyFromCache(cache, kid)
		})

		if err != nil || !token.Valid {
			utils.RespondError(c, types.HTTP.Unauthorized, "Invalid token", "INVALID_TOKEN", err.Error())
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondError(c, types.HTTP.Unauthorized, "Invalid token claims", "INVALID_CLAIMS", "Could not parse token claims")
			c.Abort()
			return
		}

		// Extract user ID
		sub, ok := claims["sub"].(string)
		if !ok {
			utils.RespondError(c, types.HTTP.Unauthorized, "Invalid token subject", "INVALID_SUBJECT", "Token subject is missing or invalid")
			c.Abort()
			return
		}

		// Validate user ID is a valid UUID
		userID, err := uuid.Parse(sub)
		if err != nil {
			utils.RespondError(c, types.HTTP.Unauthorized, "Invalid user ID in token", "INVALID_USER_ID", err.Error())
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from the context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.UUID{}, fmt.Errorf("user ID not found in context")
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("invalid user ID type in context")
	}

	return uid, nil
}

// GetUserIDString retrieves the authenticated user ID as a string
func GetUserIDString(c *gin.Context) (string, error) {
	uid, err := GetUserID(c)
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

// RequirePermission middleware checks if user has a specific permission
// This is a placeholder for RBAC integration
func RequirePermission(cache *redis.Client, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			utils.RespondError(c, types.HTTP.Unauthorized, "Authentication required", "UNAUTHORIZED", err.Error())
			c.Abort()
			return
		}

		// TODO: Implement permission check from cache/database
		// For now, just check if user exists
		ctx := context.Background()
		key := fmt.Sprintf("user:permissions:%s", userID.String())

		// Check if user permissions are cached
		exists, err := cache.Exists(ctx, key).Result()
		if err != nil || exists == 0 {
			// Permissions not cached - would fetch from DB and cache
			// For now, allow all authenticated users
		}

		c.Next()
	}
}
