package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/config"
	"github.com/ymiras/go-moderation/internal/engine"
	"github.com/ymiras/go-moderation/internal/storage"
	"go.uber.org/zap"
)

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Auth: config.AuthConfig{
			APIKeys: []string{"test-api-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	wordBank := storage.NewWordBank()
	svc := engine.NewService(cfg, wordBank, nil)

	t.Run("health endpoint bypasses auth and ratelimit", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("health endpoint returns 200", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("unauthenticated request to dify endpoint returns 401", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"point": "app.moderation.input", "params": {"query": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/dify/moderation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("unauthenticated request to standard endpoint returns 401", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"text": "hello world"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("authenticated request with valid token succeeds", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"text": "hello world", "point": "input"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}

		// Verify response has expected fields
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if _, ok := resp["flagged"]; !ok {
			t.Error("expected response to have 'flagged' field")
		}
	})

	t.Run("dify endpoint with valid token processes request", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"point": "app.moderation.input", "params": {"query": "test input"}}`
		req := httptest.NewRequest(http.MethodPost, "/dify/moderation", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("request with invalid token returns 401", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"text": "hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("request without bearer prefix returns 401", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		body := `{"text": "hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/moderate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "test-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("x-request-id header is set on response", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header to be set")
		}
	})
}

func TestAuthMiddlewareBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			APIKeys: []string{"test-api-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	wordBank := storage.NewWordBank()
	svc := engine.NewService(cfg, wordBank, nil)

	// Health endpoint should bypass auth
	t.Run("health bypasses auth", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestRateLimitMiddlewareBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			APIKeys: []string{"test-api-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	wordBank := storage.NewWordBank()
	svc := engine.NewService(cfg, wordBank, nil)

	// Health endpoint should bypass rate limit
	t.Run("health bypasses ratelimit", func(t *testing.T) {
		router := SetupRouter(cfg, svc, logger)

		// Make multiple rapid requests to confirm health doesn't get rate limited
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i, w.Code)
			}
		}
	})
}

func TestRequestIDGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			APIKeys: []string{"test-api-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	wordBank := storage.NewWordBank()
	svc := engine.NewService(cfg, wordBank, nil)

	router := SetupRouter(cfg, svc, logger)

	// Make request and check that X-Request-ID is generated
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID should not be empty")
	}

	// Verify UUID format (36 characters with hyphens)
	if len(requestID) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(requestID))
	}
}
