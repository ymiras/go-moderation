package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/dify-moderation/internal/config"
)

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid token", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(&config.AuthConfig{APIKeys: []string{"valid-key"}}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("missing authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(&config.AuthConfig{APIKeys: []string{"valid-key"}}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(&config.AuthConfig{APIKeys: []string{"valid-key"}}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic invalid-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(&config.AuthConfig{APIKeys: []string{"valid-key"}}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("health endpoint bypasses auth", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(&config.AuthConfig{APIKeys: []string{"valid-key"}}))
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
