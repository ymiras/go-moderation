#!/bin/bash
# Benchmark script for go-moderation service

set -e

# Default values
HOST="${BENCHMARK_HOST:-localhost}"
PORT="${BENCHMARK_PORT:-8080}"
ENDPOINT="${BENCHMARK_ENDPOINT:-/api/moderate}"
CONCURRENT="${BENCHMARK_CONCURRENT:-10}"
REQUESTS="${BENCHMARK_REQUESTS:-100}"

# Test payload
PAYLOAD='{"text":"Hello world","point":"input","app_id":"benchmark-test"}'

echo "=== go-moderation Benchmark ==="
echo "Host: $HOST"
echo "Port: $PORT"
echo "Endpoint: $ENDPOINT"
echo "Concurrent requests: $CONCURRENT"
echo "Total requests: $REQUESTS"
echo ""

# Function to run a single request
run_request() {
    local id="$1"
    local start=$(date +%s%N)
    curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test-api-key" \
        -d "$PAYLOAD" \
        "http://${HOST}:${PORT}${ENDPOINT}" 2>/dev/null
    local end=$(date +%s%N)
    echo $(( (end - start) / 1000000 ))
}

echo "Running benchmark..."

# Check if ab (Apache Bench) is available
if command -v ab &> /dev/null; then
    # Use Apache Bench if available
    ab -n "$REQUESTS" -c "$CONCURRENT" -p /tmp/benchmark_payload.json \
        -T "application/json" \
        -H "Authorization: Bearer test-api-key" \
        "http://${HOST}:${PORT}${ENDPOINT}"
else
    # Fallback to basic curl-based benchmark
    echo "Apache Bench not found, using curl-based benchmark"

    # Save payload to temp file
    echo "$PAYLOAD" > /tmp/benchmark_payload.json

    total_time=0
    success=0
    errors=0

    for i in $(seq 1 $REQUESTS); do
        response=$(curl -s -o /dev/null -w "%{http_code}" \
            -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-api-key" \
            --max-time 10 \
            -d "$PAYLOAD" \
            "http://${HOST}:${PORT}${ENDPOINT}" 2>/dev/null || echo "000")

        if [ "$response" = "200" ]; then
            ((success++))
        else
            ((errors++))
        fi

        if (( i % 10 == 0 )); then
            echo "Progress: $i/$REQUESTS requests completed"
        fi
    done

    echo ""
    echo "=== Results ==="
    echo "Total requests: $REQUESTS"
    echo "Successful: $success"
    echo "Errors: $errors"
    echo "Success rate: $(( success * 100 / REQUESTS ))%"
fi

echo ""
echo "Benchmark complete!"
