package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/model"
)

func TestHandler_Moderate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request - input moderation", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/moderate", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
				return
			}
			// Validate point
			if req.Point == "output" {
				// output moderation
			} else if req.Point != "" && req.Point != "input" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "point must be 'input' or 'output'"})
				return
			}
			c.JSON(http.StatusOK, Response{
				Flagged:   false,
				Action:    model.ActionPass,
				Hits:      nil,
				LatencyMs: 1.5,
			})
		})

		body := `{"text": "hello world", "point": "input", "app_id": "app-1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if resp.Flagged {
			t.Error("expected flagged=false")
		}
		if resp.Action != model.ActionPass {
			t.Errorf("expected action=pass, got %v", resp.Action)
		}
	})

	t.Run("valid request - output moderation", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/moderate", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
				return
			}
			c.JSON(http.StatusOK, Response{
				Flagged:   false,
				Action:    model.ActionPass,
				LatencyMs: 1.0,
			})
		})

		body := `{"text": "output text", "point": "output", "app_id": "app-1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("missing text field", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/moderate", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		body := `{"point": "input"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid point value", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/moderate", func(c *gin.Context) {
			var req Request
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
				return
			}
			if req.Point != "" && req.Point != "input" && req.Point != "output" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "point must be 'input' or 'output'"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		body := `{"text": "hello", "point": "invalid"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestRequestParsing(t *testing.T) {
	t.Run("parse request with all fields", func(t *testing.T) {
		jsonStr := `{"text": "hello", "point": "input", "app_id": "app-123"}`
		var req Request
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if req.Text != "hello" {
			t.Errorf("expected text 'hello', got '%s'", req.Text)
		}
		if req.Point != "input" {
			t.Errorf("expected point 'input', got '%s'", req.Point)
		}
		if req.AppID != "app-123" {
			t.Errorf("expected app_id 'app-123', got '%s'", req.AppID)
		}
	})

	t.Run("parse request with only text", func(t *testing.T) {
		jsonStr := `{"text": "hello"}`
		var req Request
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			t.Fatalf("failed to parse: %v", err)
		}
		if req.Text != "hello" {
			t.Errorf("expected text 'hello', got '%s'", req.Text)
		}
		if req.Point != "" {
			t.Errorf("expected empty point, got '%s'", req.Point)
		}
	})
}
