package main

import (
	"log"

	"github.com/ymiras/dify-moderation/internal/config"
	"github.com/ymiras/dify-moderation/internal/model"

	"github.com/gin-gonic/gin"
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

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Placeholder route for future implementation
	router.POST("/api/v1/moderate", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"flagged":    false,
			"action":     model.ActionPass,
			"hits":       []model.HitRecord{},
			"latency_ms": 0,
		})
	})

	// Start server
	logger.Info("Starting server", zap.String("addr", cfg.Server.Addr()))
	if err := router.Run(cfg.Server.Addr()); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
