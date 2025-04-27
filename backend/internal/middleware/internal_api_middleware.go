package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InternalAPIMiddleware creates a middleware for internal API authentication
func InternalAPIMiddleware(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if token is provided
		token := c.GetHeader("X-API-Token")
		if token == "" {
			// Try to get token from Authorization header
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing API token"})
			return
		}

		// Validate the token
		if token != apiToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API token"})
			return
		}

		// Check if request comes from internal network
		clientIP := c.ClientIP()
		// Allow local connections and docker subnet (simplified - in production you would check more thoroughly)
		if !strings.HasPrefix(clientIP, "127.0.0.1") && !strings.HasPrefix(clientIP, "172.") && !strings.HasPrefix(clientIP, "192.168.") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied from external network"})
			return
		}

		c.Next()
	}
}
