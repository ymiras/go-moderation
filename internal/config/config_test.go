package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Create a temp directory with no config file
	tmpDir := t.TempDir()

	// Change to temp dir so viper can't find any config
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, 8080)
	}
	if cfg.Server.ReadTimeout != 10*time.Second { // 10 seconds
		t.Errorf("Server.ReadTimeout = %v, want %v", cfg.Server.ReadTimeout, 10*time.Second)
	}

	// Test auth defaults
	if len(cfg.Auth.APIKeys) != 0 {
		t.Errorf("Auth.APIKeys = %v, want empty slice", cfg.Auth.APIKeys)
	}

	// Test rate limit defaults
	if cfg.RateLimit.Rate != 100.0 {
		t.Errorf("RateLimit.Rate = %v, want %v", cfg.RateLimit.Rate, 100.0)
	}
	if cfg.RateLimit.Capacity != 200 {
		t.Errorf("RateLimit.Capacity = %v, want %v", cfg.RateLimit.Capacity, 200)
	}

	// Test log defaults
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, "info")
	}
	if cfg.Log.Encoding != "json" {
		t.Errorf("Log.Encoding = %v, want %v", cfg.Log.Encoding, "json")
	}

	// Test moderation defaults
	if cfg.Moderation.PipelineMode != "chain" {
		t.Errorf("Moderation.PipelineMode = %v, want %v", cfg.Moderation.PipelineMode, "chain")
	}
	if cfg.Moderation.FallbackAction != "pass" {
		t.Errorf("Moderation.FallbackAction = %v, want %v", cfg.Moderation.FallbackAction, "pass")
	}

	// Test matchers defaults
	if !cfg.Matchers.AC.Enabled {
		t.Error("Matchers.AC.Enabled should be true by default")
	}
	if cfg.Matchers.Regex.Enabled {
		t.Error("Matchers.Regex.Enabled should be false by default")
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "default.yaml")

	configContent := `
server:
  host: "127.0.0.1"
  port: 9000
  read_timeout: "30s"
  write_timeout: "30s"

log:
  level: "debug"
  encoding: "console"

auth:
  api_keys:
    - "test-key-1"
    - "test-key-2"

ratelimit:
  rate: 50.0
  capacity: 100

moderation:
  pipeline_mode: "parallel"
  weighted_threshold: 0.8
  fallback_action: "block"

matchers:
  ac:
    enabled: false
  regex:
    enabled: true
  external:
    enabled: true
    api_url: "http://external.api"
    api_key: "secret"
    timeout: "10s"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create configs directory and symlink to temp
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)
	os.MkdirAll("configs", 0755)
	os.WriteFile(filepath.Join("configs", "default.yaml"), []byte(configContent), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config was loaded
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, 9000)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, "debug")
	}
	if len(cfg.Auth.APIKeys) != 2 {
		t.Errorf("len(Auth.APIKeys) = %v, want %v", len(cfg.Auth.APIKeys), 2)
	}
	if cfg.Moderation.PipelineMode != "parallel" {
		t.Errorf("Moderation.PipelineMode = %v, want %v", cfg.Moderation.PipelineMode, "parallel")
	}
	if !cfg.Matchers.Regex.Enabled {
		t.Error("Matchers.Regex.Enabled should be true")
	}
	if !cfg.Matchers.External.Enabled {
		t.Error("Matchers.External.Enabled should be true")
	}
	if cfg.Matchers.External.APIURL != "http://external.api" {
		t.Errorf("Matchers.External.APIURL = %v, want %v", cfg.Matchers.External.APIURL, "http://external.api")
	}
}

func TestServerConfig_Addr(t *testing.T) {
	cfg := &ServerConfig{
		Host: "192.168.1.1",
		Port: 8888,
	}

	expected := "192.168.1.1:8888"
	if got := cfg.Addr(); got != expected {
		t.Errorf("Addr() = %v, want %v", got, expected)
	}
}
