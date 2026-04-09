package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/config"
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
