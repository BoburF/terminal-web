.PHONY: all build run run-server start-server stop-server restart-server status test-ssh clean deps fmt lint help setup gen-key tail-logs stats check-security

# Variables
BINARY_NAME := terminal-web
PORT := 22
LOG_FILE := logs/terminal-web.log
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Default target
all: build

# Create necessary directories
setup:
	@mkdir -p keys logs
	@echo "Setup complete. Run 'make gen-key' to generate host key if not already done."

# Generate host key
gen-key:
	@if [ ! -f keys/ssh_host_ed25519_key ]; then \
		ssh-keygen -t ed25519 -f keys/ssh_host_ed25519_key -N "" -C "terminal-web-host-key"; \
		chmod 600 keys/ssh_host_ed25519_key; \
		chmod 644 keys/ssh_host_ed25519_key.pub; \
		echo "Host key generated successfully"; \
	else \
		echo "Host key already exists"; \
	fi

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
run-server: build setup
	@echo "=========================================="
	@echo "  SSH Server Starting on port $(PORT)"
	@echo "=========================================="
	@echo ""
	@echo "Security Features:"
	@echo "  - SSH key authentication only (no passwords)"
	@echo "  - Max 30 concurrent connections"
	@echo "  - Rate limit: 10 connections/minute per IP"
	@echo "  - Session timeout: 10 minutes max"
	@echo ""
	@echo "Logs written to: $(LOG_FILE)"
	@echo ""
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
start-server: build setup
	@echo "Starting SSH server on port $(PORT) in background..."
	@./$(BINARY_NAME) -server -port $(PORT) >> $(LOG_FILE) 2>&1 &
	@sleep 1
	@pid=$$(pgrep -f "$(BINARY_NAME) -server.*$(PORT)" | head -1); \
	if [ -n "$$pid" ]; then \
		echo "SSH server started. PID: $$pid"; \
		echo "Connect with: ssh -p $(PORT) localhost"; \
		echo "View logs: tail -f $(LOG_FILE)"; \
	else \
		echo "Failed to start server. Check logs: $(LOG_FILE)"; \
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

# View logs
tail-logs:
	@if [ -f $(LOG_FILE) ]; then \
		tail -f $(LOG_FILE); \
	else \
		echo "Log file not found: $(LOG_FILE)"; \
	fi

# Show connection stats
stats:
	@echo "=== Server Statistics ==="
	@echo "Active processes:"
	@ps aux | grep "$(BINARY_NAME) -server" | grep -v grep | wc -l | xargs echo "  Running instances:"
	@echo ""
	@echo "Recent connections (last 10):"
	@if [ -f $(LOG_FILE) ]; then \
		tail -100 $(LOG_FILE) | grep '"event":"CONNECT"' | tail -10 | jq -r '[.timestamp, .ip, .key_fingerprint[:20]] | @tsv' 2>/dev/null || echo "  (Install jq for better formatting)"; \
	else \
		echo "  No log file found"; \
	fi
	@echo ""
	@echo "Total connections today:"
	@if [ -f $(LOG_FILE) ]; then \
		grep $$(date +%Y-%m-%d) $(LOG_FILE) | grep '"event":"CONNECT"' | wc -l | xargs echo "  "; \
	else \
		echo "  No log file found"; \
	fi
	@echo ""
	@echo "Rate limited connections (last hour):"
	@if [ -f $(LOG_FILE) ]; then \
		grep '"event":"RATE_LIMIT"' $(LOG_FILE) | tail -20 | wc -l | xargs echo "  "; \
	else \
		echo "  No log file found"; \
	fi

# Security check
check-security:
	@echo "=== Security Configuration Check ==="
	@echo ""
	@echo "Host Key:"
	@if [ -f keys/ssh_host_ed25519_key ]; then \
		echo "  ✓ Host key exists"; \
		key_perms=$$(stat -c %a keys/ssh_host_ed25519_key); \
		if [ "$$key_perms" = "600" ]; then \
			echo "  ✓ Host key permissions correct (600)"; \
		else \
			echo "  ✗ Host key permissions incorrect (expected 600, got $$key_perms)"; \
		fi; \
	else \
		echo "  ✗ Host key missing - run 'make gen-key'"; \
	fi
	@echo ""
	@echo "Log Directory:"
	@if [ -d logs ]; then \
		echo "  ✓ Log directory exists"; \
		if [ -f $(LOG_FILE) ]; then \
			echo "  ✓ Log file exists"; \
		else \
			echo "  ⚠ Log file does not exist yet (will be created on first run)"; \
		fi; \
	else \
		echo "  ✗ Log directory missing - run 'make setup'"; \
	fi
	@echo ""
	@echo "Firewall:"
	@if command -v ufw >/dev/null 2>&1; then \
		if sudo ufw status | grep -q "$(PORT)"; then \
			echo "  ✓ Firewall rule exists for port $(PORT)"; \
		else \
			echo "  ⚠ No firewall rule found for port $(PORT)"; \
			echo "    Run: sudo ufw allow $(PORT)/tcp"; \
		fi; \
	elif command -v iptables >/dev/null 2>&1; then \
		echo "  ⚠ iptables detected - ensure port $(PORT) is allowed manually"; \
	else \
		echo "  ⚠ No firewall detected (ufw/iptables)"; \
	fi
	@echo ""
	@echo "Binary:"
	@if [ -f $(BINARY_NAME) ]; then \
		echo "  ✓ Binary exists: $(BINARY_NAME)"; \
	else \
		echo "  ✗ Binary not found - run 'make build'"; \
	fi
	@echo ""
	@echo "Resume Files:"
	@if [ -f resume/index.html ]; then \
		echo "  ✓ Resume HTML exists"; \
	else \
		echo "  ✗ Resume HTML missing"; \
	fi
	@echo ""
	@echo "Configuration Summary:"
	@echo "  Port: $(PORT)"
	@echo "  Max Connections: 30"
	@echo "  Rate Limit: 10/min per IP"
	@echo "  Session Timeout: 10 minutes"
	@echo "  Authentication: SSH key only"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "Clean complete"

# Clean everything including logs and keys
clean-all: clean
	@echo "Removing logs and keys..."
	rm -f logs/*.log
	rm -f keys/ssh_host_*
	@echo "Clean all complete"

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
	@echo "  make setup            - Create necessary directories"
	@echo "  make gen-key          - Generate SSH host key"
	@echo "  make build            - Build the binary"
	@echo "  make run              - Run in local mode"
	@echo "  make run-server       - Run SSH server (foreground, blocks terminal)"
	@echo "  make start-server     - Start SSH server in background"
	@echo "  make stop-server      - Stop background SSH server"
	@echo "  make restart-server   - Restart SSH server"
	@echo "  make status           - Check if server is running"
	@echo "  make test-ssh         - Test SSH connection"
	@echo "  make tail-logs        - View live logs"
	@echo "  make stats            - Show connection statistics"
	@echo "  make check-security   - Verify security configuration"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make clean-all        - Clean everything including logs/keys"
	@echo "  make deps             - Install dependencies"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Run linter"
	@echo "  make help             - Show this help message"
	@echo ""
	@echo "Quick Start:"
	@echo "  make gen-key && make run-server          # Generate key and start"
	@echo "  make check-security                      # Verify everything is set up"
