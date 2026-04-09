package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/ymiras/go-moderation/internal/api/dify"
	"github.com/ymiras/go-moderation/internal/api/middleware"
	standardadapter "github.com/ymiras/go-moderation/internal/api/standard"
	"github.com/ymiras/go-moderation/internal/config"
	"github.com/ymiras/go-moderation/internal/engine"

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
	standardHandler := standardadapter.NewHandler(svc)
	r.POST("/api/v1/text/moderation", standardHandler.Moderate)

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

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Info("Server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	log.Info("Received signal, initiating graceful shutdown", zap.String("signal", sig.String()))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Warn("Server shutdown error", zap.Error(err))
		return err
	}

	log.Info("Server shutdown complete")
	return nil
}
