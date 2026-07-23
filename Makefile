# Makefile for URL Shortener
# Run 'make help' to see all available commands

.PHONY: help build run test clean docker-up docker-down docker-logs migrate-up migrate-down

# Default target
help:
	@echo "Available commands:"
	@echo "  make build        - Build the Go binary"
	@echo "  make run          - Run the server locally (requires Postgres & Redis running)"
	@echo "  make test         - Run all unit tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make docker-up    - Start all services with Docker Compose"
	@echo "  make docker-down  - Stop all Docker Compose services"
	@echo "  make docker-logs  - View logs from Docker Compose services"
	@echo "  make fmt          - Format Go code"
	@echo "  make vet          - Run go vet for code analysis"

# Build the application binary
build:
	go build -o bin/server ./cmd/server

# Run the server (requires local Postgres and Redis)
run:
	go run ./cmd/server

# Run all tests with verbose output
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f server

# Start everything with Docker Compose
docker-up:
	docker-compose up --build -d

# Stop Docker Compose services
docker-down:
	docker-compose down

# View Docker Compose logs
docker-logs:
	docker-compose logs -f

# Format Go code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...