.PHONY: build run test clean docker-build docker-run docker-compose-up docker-compose-down lint

# Variables
BINARY_NAME=dify-moderation
BINARY_DIR=bin
BUILD_DIR=.
DOCKER_IMAGE=dify-moderation:latest
COMPOSE_FILE=docker-compose.yml

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(BUILD_DIR)/cmd/server

# Run the service
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f $(BINARY_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Docker build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --rm $(DOCKER_IMAGE)

# Docker compose up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose -f $(COMPOSE_FILE) up -d

# Docker compose down
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose -f $(COMPOSE_FILE) down

# Docker compose build
docker-compose-build:
	@echo "Building services with docker-compose..."
	docker-compose -f $(COMPOSE_FILE) build

# Build binary for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_DIR)/$(BINARY_NAME) $(BUILD_DIR)/cmd/server

# Help
help:
	@echo "Available targets:"
	@echo "  build             - Build the binary"
	@echo "  run               - Build and run the service"
	@echo "  test              - Run tests"
	@echo "  clean             - Clean build artifacts"
	@echo "  deps              - Download dependencies"
	@echo "  lint              - Run linter"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container"
	@echo "  docker-compose-up    - Start services with docker-compose"
	@echo "  docker-compose-down  - Stop services with docker-compose"
	@echo "  build-linux       - Build binary for Linux"
