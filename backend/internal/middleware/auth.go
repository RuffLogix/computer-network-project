package middleware

import (
	"net/http"
	"strconv"
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

// ChatMembershipMiddleware checks if user is a member of private chats
func ChatMembershipMiddleware(chatService service.ChatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("id")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDInt := userID.(int64)

		// Get chat details
		chat, err := chatService.GetChat(chatID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
			c.Abort()
			return
		}

		// Allow access to public chats
		if chat.IsPublic {
			c.Next()
			return
		}

		// Check if user is a member of private chat
		members, err := chatService.GetMembers(chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
			c.Abort()
			return
		}

		isMember := false
		for _, member := range members {
			if member.UserID == userIDInt {
				isMember = true
				break
			}
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: not a member of this private chat"})
			c.Abort()
			return
		}

		c.Next()
	}
}
