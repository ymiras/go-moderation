package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/ymiras/go-moderation/internal/config"
)

// RateLimit returns a middleware that implements token bucket rate limiting.
func RateLimit(rateLimitCfg *config.RateLimitConfig) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateLimitCfg.Rate), rateLimitCfg.Capacity)

	return func(c *gin.Context) {
		// Skip rate limit for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		if !limiter.Allow() {
			retryAfter := 1
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			return
		}

		c.Next()
	}
}
