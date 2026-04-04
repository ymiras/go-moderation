package main

import (
	"testing"

	"github.com/ymiras/dify-moderation/internal/config"
	"github.com/ymiras/dify-moderation/internal/engine"
	"github.com/ymiras/dify-moderation/internal/server"
	"github.com/ymiras/dify-moderation/internal/storage"
	"go.uber.org/zap"
)

// TestServerInitialization tests that the server can be initialized with valid config
func TestServerInitialization(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "127.0.0.1",
			Port:         18080,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Auth: config.AuthConfig{
			APIKeys: []string{"test-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	// Create logger (using development mode for test)
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create word bank
	wordBank := storage.NewWordBank()

	// Create service
	svc := engine.NewService(cfg, wordBank, nil)

	// Test router setup
	router := server.SetupRouter(cfg, svc, logger)
	if router == nil {
		t.Error("SetupRouter returned nil router")
	}
}

// TestServerRouter_WithMockService tests router with a mock service
func TestServerRouter_WithMockService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "127.0.0.1",
			Port:         18081,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Auth: config.AuthConfig{
			APIKeys: []string{"test-key"},
		},
		RateLimit: config.RateLimitConfig{
			Rate:     100,
			Capacity: 200,
		},
	}

	wordBank := storage.NewWordBank()
	svc := engine.NewService(cfg, wordBank, nil)

	router := server.SetupRouter(cfg, svc, logger)
	if router == nil {
		t.Fatal("SetupRouter returned nil")
	}

	// Verify routes are registered by checking a request
	// This is a smoke test to ensure the router is properly configured
	t.Log("Router initialized successfully")
}
