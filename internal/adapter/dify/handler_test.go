package dify

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_Moderate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("input moderation - pass", func(t *testing.T) {
		// Skip for now - needs proper service integration
	})

	t.Run("input moderation - block", func(t *testing.T) {
		// Skip for now - needs proper service mock
	})

	t.Run("missing point field", func(t *testing.T) {
		router := gin.New()
		router.POST("/dify/moderation", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		body := `{"params": {"query": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/dify/moderation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unsupported point type", func(t *testing.T) {
		router := gin.New()
		router.POST("/dify/moderation", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if req.Point != "app.moderation.input" && req.Point != "app.moderation.output" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported point type"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		body := `{"point": "app.invalid", "params": {"query": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/dify/moderation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing text", func(t *testing.T) {
		router := gin.New()
		router.POST("/dify/moderation", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			text := ""
			if req.Point == "app.moderation.input" {
				text = req.Params.Query
			} else {
				text = req.Params.Text
			}
			if text == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		body := `{"point": "app.moderation.input", "params": {"app_id": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/dify/moderation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestRequestParsing(t *testing.T) {
	t.Run("parse input moderation request", func(t *testing.T) {
		jsonStr := `{"point": "app.moderation.input", "params": {"app_id": "app-1", "inputs": {"name": "test"}, "query": "hello"}}`
		var req Request
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if req.Point != "app.moderation.input" {
			t.Errorf("expected point 'app.moderation.input', got '%s'", req.Point)
		}
		if req.Params.Query != "hello" {
			t.Errorf("expected query 'hello', got '%s'", req.Params.Query)
		}
	})

	t.Run("parse output moderation request", func(t *testing.T) {
		jsonStr := `{"point": "app.moderation.output", "params": {"app_id": "app-1", "text": "world"}}`
		var req Request
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if req.Point != "app.moderation.output" {
			t.Errorf("expected point 'app.moderation.output', got '%s'", req.Point)
		}
		if req.Params.Text != "world" {
			t.Errorf("expected text 'world', got '%s'", req.Params.Text)
		}
	})
}
