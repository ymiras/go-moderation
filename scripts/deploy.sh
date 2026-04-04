#!/bin/bash
# Deploy script for dify-moderation service using Docker

set -e

IMAGE_NAME="${IMAGE_NAME:-dify-moderation}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_NAME="${CONTAINER_NAME:-dify-moderation}"
HOST_PORT="${HOST_PORT:-8080}"

echo "=== dify-moderation Deploy Script ==="
echo "Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "Container name: ${CONTAINER_NAME}"
echo "Host port: ${HOST_PORT}"
echo ""

# Stop and remove existing container if it exists
echo "Cleaning up existing container..."
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

# Build the image
echo "Building Docker image..."
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Run the container
echo "Starting container..."
docker run -d \
    --name "$CONTAINER_NAME" \
    -p "${HOST_PORT}:8080" \
    -v "$(pwd)/configs:/app/configs:ro" \
    --restart unless-stopped \
    "${IMAGE_NAME}:${IMAGE_TAG}"

echo ""
echo "Container started. Waiting for service to be healthy..."

# Wait for health check
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    if docker exec "$CONTAINER_NAME" wget --no-verbose --tries=1 --spider http://localhost:8080/health 2>/dev/null; then
        echo ""
        echo "Service is healthy!"
        echo "Deployment complete. Service is running at http://localhost:${HOST_PORT}"
        exit 0
    fi
    ((attempt++))
    echo "Waiting for service to be ready... ($attempt/$max_attempts)"
    sleep 2
done

echo ""
echo "Warning: Health check did not pass, but container is running."
echo "Deployment complete. Service is running at http://localhost:${HOST_PORT}"
echo ""
echo "To check logs: docker logs $CONTAINER_NAME"
echo "To stop: docker stop $CONTAINER_NAME"
