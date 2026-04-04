# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o dify-moderation ./cmd/server

# Final stage - minimal image
FROM alpine:3.19

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/dify-moderation .
COPY --from=builder /build/configs ./configs

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./dify-moderation"]
