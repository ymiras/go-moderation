package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/config"
)

// RateLimit returns a middleware that implements token bucket rate limiting.
func RateLimit(rateLimitCfg *config.RateLimitConfig) gin.HandlerFunc {
	bucket := newTokenBucket(rateLimitCfg.Rate, rateLimitCfg.Capacity)

	return func(c *gin.Context) {
		// Skip rate limit for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		if !bucket.Allow() {
			retryAfter := int(time.Second.Seconds()) // 1 second
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			c.Header("Retry-After", string(rune(retryAfter)))
			return
		}

		c.Next()
	}
}

// tokenBucket implements the token bucket algorithm.
type tokenBucket struct {
	rate     float64
	capacity int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// newTokenBucket creates a new token bucket with the given rate and capacity.
func newTokenBucket(rate float64, capacity int) *tokenBucket {
	return &tokenBucket{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		lastTime: time.Now(),
	}
}

// Allow checks if a request is allowed.
func (b *tokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime).Seconds()
	b.lastTime = now

	// Add tokens based on elapsed time
	b.tokens += elapsed * b.rate
	if b.tokens > float64(b.capacity) {
		b.tokens = float64(b.capacity)
	}

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}
