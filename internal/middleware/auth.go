package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/dify-moderation/internal/config"
)

// Auth returns a middleware that validates Bearer tokens.
func Auth(authCfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
			})
			return
		}

		token := parts[1]
		if !isValidAPIKey(token, authCfg.APIKeys) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Next()
	}
}

// isValidAPIKey checks if the given token matches any of the configured API keys.
func isValidAPIKey(token string, apiKeys []string) bool {
	for _, key := range apiKeys {
		if key == token {
			return true
		}
	}
	return false
}
