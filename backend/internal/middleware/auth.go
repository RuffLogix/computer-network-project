package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rufflogix/computer-network-project/internal/service"
)

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.NumericID)    // Numeric ID for chat operations
		c.Set("user_id_str", user.ID.Hex()) // Store as string for MongoDB
		c.Set("user_object", user)          // Store full user object
		c.Set("is_guest", user.IsGuest)
		c.Next()
	}
}

// OptionalAuthMiddleware allows both authenticated and guest access
func OptionalAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// If no header token, try query parameter (for WebSocket)
		if token == "" {
			token = c.Query("token")
		}

		// Validate token if present
		if token != "" {
			user, err := authService.ValidateToken(token)
			if err == nil {
				c.Set("user", user)
				c.Set("user_id", user.NumericID)
				c.Set("user_id_str", user.ID.Hex())
				c.Set("user_object", user)
				c.Set("is_guest", user.IsGuest)
			}
		}

		c.Next()
	}
}

// GuestOrAuthMiddleware allows guest users or authenticated users
func GuestOrAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, consider as guest
			c.Set("is_guest", true)
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id_str", user.ID.Hex())
		c.Set("user_object", user)
		c.Set("is_guest", user.IsGuest)
		c.Next()
	}
}
