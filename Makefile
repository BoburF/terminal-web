.PHONY: all build run run-server start-server stop-server restart-server status test-ssh clean deps fmt lint help

# Variables
BINARY_NAME := terminal-web.git
PORT := 4444
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)
	@echo "Build complete: ./$(BINARY_NAME)"

# Run in local mode (requires terminal)
run: build
	@echo "Running in local mode..."
	./$(BINARY_NAME)

# Run SSH server (foreground - blocks terminal, shows logs live)
run-server: build
	@echo "=========================================="
	@echo "  SSH Server Starting on port $(PORT)"
	@echo "=========================================="
	@echo ""
	@echo "Server is running and waiting for connections..."
	@echo "Connect from another computer:"
	@echo "  ssh -p $(PORT) <your-ip-address>"
	@echo ""
	@echo "Press Ctrl+C to stop the server"
	@echo "=========================================="
	@echo ""
	./$(BINARY_NAME) -server -port $(PORT)

# Restart SSH server
restart-server: stop-server
	@sleep 1
	@$(MAKE) start-server

# Run SSH server in background
start-server: build
	@echo "Starting SSH server on port $(PORT) in background..."
	@./$(BINARY_NAME) -server -port $(PORT) > /tmp/terminal-web.log 2>&1 &
	@sleep 1
	@pid=$$(pgrep -f "$(BINARY_NAME) -server.*$(PORT)" | head -1); \
	if [ -n "$$pid" ]; then \
		echo "SSH server started. PID: $$pid"; \
		echo "Connect with: ssh -p $(PORT) localhost"; \
		echo "View logs: tail -f /tmp/terminal-web.log"; \
	else \
		echo "Failed to start server. Check logs: /tmp/terminal-web.log"; \
	fi

# Check server status
status:
	@pid=$$(pgrep -f "$(BINARY_NAME) -server.*$(PORT)" | head -1); \
	if [ -n "$$pid" ]; then \
		echo "SSH server is running. PID: $$pid, Port: $(PORT)"; \
	else \
		echo "SSH server is not running"; \
	fi

# Stop background SSH server
stop-server:
	@echo "Stopping SSH server..."
	@pid=$$(pgrep -f "$(BINARY_NAME) -server.*$(PORT)" | head -1); \
	if [ -n "$$pid" ]; then \
		kill $$pid 2>/dev/null && echo "SSH server stopped (PID: $$pid)" || echo "Failed to stop server"; \
	else \
		echo "SSH server is not running"; \
	fi

# Test SSH connection (requires server to be running)
test-ssh:
	@echo "Testing SSH connection..."
	ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $(PORT) localhost

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  make build         - Build the binary"
	@echo "  make run           - Run in local mode"
	@echo "  make run-server    - Run SSH server (foreground, blocks terminal)"
	@echo "  make start-server  - Start SSH server in background"
	@echo "  make stop-server   - Stop background SSH server"
	@echo "  make restart-server- Restart SSH server"
	@echo "  make status        - Check if server is running"
	@echo "  make test-ssh      - Test SSH connection"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Install dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run                     # Run locally"
	@echo "  make start-server            # Start server in background"
	@echo "  make status                  # Check server status"
	@echo "  ssh -p 4444 localhost        # Connect to SSH server"
	@echo "  make stop-server             # Stop the server"
