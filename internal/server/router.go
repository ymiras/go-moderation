package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/dify-moderation/internal/adapter/dify"
	"github.com/ymiras/dify-moderation/internal/adapter/standard"
	"github.com/ymiras/dify-moderation/internal/config"
	"github.com/ymiras/dify-moderation/internal/engine"
	"github.com/ymiras/dify-moderation/internal/middleware"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the Gin router.
func SetupRouter(cfg *config.Config, svc *engine.ModerationService, log *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middleware order: Logger → Auth → RateLimit → Recovery
	r.Use(middleware.Logger(log))
	r.Use(middleware.Auth(&cfg.Auth))
	r.Use(middleware.RateLimit(&cfg.RateLimit))
	r.Use(gin.Recovery())

	// Health check endpoint (no auth/ratelimit required - handled in middleware)
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Dify adapter
	difyHandler := dify.NewHandler(svc)
	r.POST("/dify/moderation", difyHandler.Moderate)

	// Standard adapter
	standardHandler := standard.NewHandler(svc)
	r.POST("/api/moderate", standardHandler.Moderate)

	return r
}

// StartServer starts the HTTP server with graceful shutdown.
func StartServer(cfg *config.Config, svc *engine.ModerationService, log *zap.Logger) error {
	r := SetupRouter(cfg, svc, log)

	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// TODO: Implement graceful shutdown with signal handling
	// For now, just start the server
	return srv.ListenAndServe()
}
