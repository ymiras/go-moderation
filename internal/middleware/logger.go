package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Logger returns a middleware that logs requests in structured JSON format.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start).Seconds()

		// Log request
		log.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Float64("latency_ms", latency*1000),
			zap.String("request_id", requestID),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
