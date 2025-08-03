# Microservice Project Makefile

.PHONY: help build run test clean install-deps

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build all services"
	@echo "  run-auth     - Run auth service"
	@echo "  run-file     - Run file service"
	@echo "  run-user     - Run user service"
	@echo "  run-analytics - Run analytics service"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-deps - Install Go dependencies"
	@echo "  lint         - Run linter"

# Build all services
build:
	@echo "Building all services..."
	go build -o bin/auth ./services/auth
	go build -o bin/file ./services/file
	go build -o bin/user ./services/user
	go build -o bin/analytics ./services/analytics
	@echo "Build complete!"

# Run individual services
run-auth:
	@echo "Starting auth service on port 8081..."
	SERVER_PORT=8081 go run ./services/auth

run-file:
	@echo "Starting file service on port 8082..."
	SERVER_PORT=8082 go run ./services/file

run-user:
	@echo "Starting user service on port 8083..."
	SERVER_PORT=8083 go run ./services/user

run-analytics:
	@echo "Starting analytics service on port 8085..."
	SERVER_PORT=8085 go run ./services/analytics

# Test commands
test:
	@echo "Running all tests..."
	go test ./...

test-unit:
	@echo "Running unit tests..."
	go test ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/...

# Development commands
install-deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# Build with bin directory creation
build: bin