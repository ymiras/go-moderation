package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets request ID header", func(t *testing.T) {
		log, _ := zap.NewDevelopment()
		router := gin.New()
		router.Use(Logger(log))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header to be set")
		}
	})

	t.Run("sets request ID in context", func(t *testing.T) {
		log, _ := zap.NewDevelopment()
		var capturedID string
		router := gin.New()
		router.Use(Logger(log))
		router.GET("/test", func(c *gin.Context) {
			if id, exists := c.Get("request_id"); exists {
				capturedID = id.(string)
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if capturedID == "" {
			t.Error("expected request_id to be set in context")
		}
	})

	t.Run("logs request", func(t *testing.T) {
		log, _ := zap.NewDevelopment()
		router := gin.New()
		router.Use(Logger(log))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Just verify no panic and request completes
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}
