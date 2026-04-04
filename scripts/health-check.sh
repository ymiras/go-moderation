#!/bin/bash
# Health check script for dify-moderation service

set -e

# Default values
HOST="${HEALTH_HOST:-localhost}"
PORT="${HEALTH_PORT:-8080}"
TIMEOUT="${HEALTH_TIMEOUT:-5}"

# Function to check health
check_health() {
    local host="$1"
    local port="$2"
    local timeout="$3"

    # Try to get HTTP response code
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        --max-time "$timeout" \
        "http://${host}:${port}/health" 2>/dev/null || echo "000")

    if [ "$response" = "200" ]; then
        echo "Service is healthy"
        return 0
    else
        echo "Service is unhealthy (HTTP status: $response)"
        return 1
    fi
}

# Main
echo "Checking health of dify-moderation at ${HOST}:${PORT}..."

if check_health "$HOST" "$PORT" "$TIMEOUT"; then
    exit 0
else
    exit 1
fi
