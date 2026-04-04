package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/dify-moderation/internal/config"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows requests within limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimit(&config.RateLimitConfig{Rate: 10, Capacity: 10}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// First request should succeed
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("health endpoint bypasses rate limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimit(&config.RateLimitConfig{Rate: 0, Capacity: 0})) // rate 0 would block everything
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, nil)
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestTokenBucket(t *testing.T) {
	t.Run("allows burst up to capacity", func(t *testing.T) {
		bucket := newTokenBucket(1, 5) // 1 token per second, capacity 5

		// Should allow 5 requests immediately
		for i := 0; i < 5; i++ {
			if !bucket.Allow() {
				t.Errorf("request %d should be allowed", i+1)
			}
		}
	})

	t.Run("blocks when empty", func(t *testing.T) {
		bucket := newTokenBucket(0, 0) // no capacity

		if bucket.Allow() {
			t.Error("request should be blocked")
		}
	})

	t.Run("refills over time", func(t *testing.T) {
		bucket := newTokenBucket(100, 1) // 100 tokens per second, capacity 1

		// Use the only token
		if !bucket.Allow() {
			t.Error("first request should be allowed")
		}

		// Should be empty now
		if bucket.Allow() {
			t.Error("second request should be blocked immediately")
		}

		// Wait for refill
		time.Sleep(20 * time.Millisecond) // 20ms * 100 = 2 tokens

		if !bucket.Allow() {
			t.Error("request should be allowed after refill")
		}
	})
}
