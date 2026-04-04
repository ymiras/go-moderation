package main

import (
	"log"

	"github.com/ymiras/dify-moderation/internal/config"
	"github.com/ymiras/dify-moderation/internal/engine"
	"github.com/ymiras/dify-moderation/internal/matcher"
	"github.com/ymiras/dify-moderation/internal/server"
	"github.com/ymiras/dify-moderation/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize word bank
	wordBank := storage.NewWordBank()
	if err := wordBank.Load("configs/wordlist/default.csv"); err != nil {
		logger.Warn("Failed to load word bank, using empty word bank", zap.Error(err))
	}

	// Initialize matchers based on config
	var matchers []matcher.Matcher
	if cfg.Matchers.AC.Enabled {
		m, err := matcher.NewAC(&matcher.ACConfig{})
		if err != nil {
			logger.Fatal("Failed to create AC matcher", zap.Error(err))
		}
		matchers = append(matchers, m)
		logger.Info("AC matcher enabled")
	}
	if cfg.Matchers.Regex.Enabled {
		m, err := matcher.NewRegex(&matcher.RegexConfig{})
		if err != nil {
			logger.Fatal("Failed to create Regex matcher", zap.Error(err))
		}
		matchers = append(matchers, m)
		logger.Info("Regex matcher enabled")
	}
	if cfg.Matchers.External.Enabled {
		m, err := matcher.NewExternal(&matcher.ExternalConfig{
			Endpoint: cfg.Matchers.External.APIURL,
			APIKey:   cfg.Matchers.External.APIKey,
			Timeout:  cfg.Matchers.External.Timeout,
		})
		if err != nil {
			logger.Fatal("Failed to create External matcher", zap.Error(err))
		}
		matchers = append(matchers, m)
		logger.Info("External matcher enabled", zap.String("api_url", cfg.Matchers.External.APIURL))
	}

	// Create moderation service
	svc := engine.NewService(cfg, wordBank, matchers)

	// Setup router and start server
	if err := server.StartServer(cfg, svc, logger); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
