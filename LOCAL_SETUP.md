# Local SSH Server Setup Guide

Complete guide for running the Terminal-Web SSH resume server locally on your computer.

## Prerequisites

### System Requirements
- **OS**: Linux, macOS, or WSL (Windows Subsystem for Linux)
- **Go**: Version 1.23 or higher
- **SSH Client**: OpenSSH (usually pre-installed)
- **Git**: For cloning the repository (optional)

### Check Prerequisites

```bash
# Check Go version
go version

# Check SSH client
ssh -V

# Check Git
git --version
```

## Installation

### 1. Clone or Navigate to Repository

```bash
# If cloning from git
git clone <repository-url>
cd terminal-web

# Or if already in the directory
cd /path/to/terminal-web
```

### 2. Install Dependencies

```bash
# Download and install Go dependencies
make deps
```

This will run:
- `go mod tidy` - Clean up module dependencies
- `go mod download` - Download required packages

### 3. Build the Application

```bash
# Compile the binary
make build
```

This creates a `terminal-web` binary in the current directory.

## Initial Setup

### 1. Create Directories

```bash
# Create keys and logs directories
make setup
```

This creates:
- `keys/` - For SSH host key storage
- `logs/` - For audit logs

### 2. Generate SSH Host Key

```bash
# Generate Ed25519 host key (one-time setup)
make gen-key
```

This creates:
- `keys/ssh_host_ed25519_key` (private key, 600 permissions)
- `keys/ssh_host_ed25519_key.pub` (public key)

**Important**: Keep the private key secure. Never commit it to git.

### 3. Verify Setup

```bash
# Check security configuration
make check-security
```

Expected output:
```
=== Security Configuration Check ===

Host Key:
  ✓ Host key exists
  ✓ Host key permissions correct (600)

Log Directory:
  ✓ Log directory exists
  ⚠ Log file does not exist yet (will be created on first run)

Binary:
  ✓ Binary exists: terminal-web

Resume Files:
  ✓ Resume HTML exists
```

## Running the SSH Server

### Method 1: Foreground Mode (Recommended for Testing)

Run the server in your current terminal. You'll see logs in real-time.

```bash
# Start server in foreground (default port 4569)
make run-server
```

You'll see output like:
```
==========================================
  SSH Server Starting on port 4569
==========================================

Security Features:
  - SSH key authentication only (no passwords)
  - Max 30 concurrent connections
  - Rate limit: 10 connections/minute per IP
  - Session timeout: 10 minutes max

Logs written to: logs/terminal-web.log

Connect from another computer:
  ssh -p 4569 <your-ip-address>

Press Ctrl+C to stop the server
==========================================

2026/02/15 10:30:00 Starting SSH server on 0.0.0.0:4569...
2026/02/15 10:30:00 Security: max 30 connections, 10/min rate limit
2026/02/15 10:30:00 Session limits: 5m0s idle timeout, 10m0s max duration
```

**To stop**: Press `Ctrl+C`

### Method 2: Background Mode

Run the server in the background (detached from terminal).

```bash
# Start server in background
make start-server
```

Check status:
```bash
make status
```

View logs:
```bash
make tail-logs
```

Stop server:
```bash
make stop-server
```

### Method 3: Custom Port

```bash
# Run on custom port (e.g., 7022)
make run-server PORT=7022
```

## Testing Locally

### Test 1: Check Server is Running

```bash
# Check if server process exists
make status

# Or use ps
ps aux | grep terminal-web
```

### Test 2: Connect via SSH

**First, ensure you have an SSH key**:

```bash
# Check if you have an SSH key
ls ~/.ssh/id_*

# If not, generate one
ssh-keygen -t ed25519 -C "your-email@example.com"
```

**Connect to the server**:

```bash
# Connect locally (skip host key checking for testing)
ssh -p 4569 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

**What you should see**:
1. The server's TUI (Terminal User Interface) with your resume
2. Navigation instructions at the bottom
3. Your resume sections displayed

**Navigation**:
- `Tab` - Next section
- `Shift+Tab` - Previous section
- `j` or `↓` - Scroll down
- `k` or `↑` - Scroll up
- `q` or `Ctrl+C` - Exit

### Test 3: Verify Logs

In another terminal window:

```bash
# View connection logs
cat logs/terminal-web.log

# Or follow live logs
tail -f logs/terminal-web.log
```

You should see JSON entries like:
```json
{"timestamp":"2026-02-15T10:35:00Z","event":"CONNECT","ip":"127.0.0.1","key_fingerprint":"SHA256:abc123...","message":"New SSH connection established"}
```

### Test 4: Test Rate Limiting

Try connecting multiple times rapidly:

```bash
# This should work (within rate limit)
ssh -p 4569 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost

# Try 15 times rapidly (should trigger rate limit)
for i in {1..15}; do
  ssh -p 4569 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost &
done
```

You should see "Rate limit exceeded" message after 10 connections.

## Local Network Testing

### Find Your IP Address

```bash
# Linux
hostname -I

# macOS
ifconfig | grep "inet " | grep -v 127.0.0.1

# Or use ip command
ip addr show
```

### Connect from Another Device

From another computer on the same network:

```bash
ssh -p 4569 -o StrictHostKeyChecking=no <your-local-ip>
```

Example:
```bash
ssh -p 4569 -o StrictHostKeyChecking=no 192.168.1.100
```

## Development Workflow

### Making Changes to Resume

1. Edit `resume/index.html`
2. Restart server (changes take effect immediately on new connections):

```bash
# If running in foreground: Ctrl+C, then:
make run-server

# If running in background:
make restart-server
```

### Rebuilding After Code Changes

```bash
# Clean and rebuild
make clean
make build

# Or just rebuild
make build
```

### Viewing Statistics

```bash
# Show connection stats
make stats
```

Example output:
```
=== Server Statistics ===
Active processes:
  Running instances: 1

Recent connections (last 10):
  2026-02-15T10:30:00Z  127.0.0.1  SHA256:abc123...
  2026-02-15T10:35:00Z  192.168.1.5  SHA256:def456...

Total connections today:
  15

Rate limited connections (last hour):
  3
```

## Troubleshooting

### "Connection refused"

**Problem**: Server is not running or port is blocked.

**Solution**:
```bash
# Check if server is running
make status

# If not running, start it
make run-server

# Check if port is in use
sudo lsof -i :4569

# Kill process using port if needed
sudo kill -9 <PID>
```

### "Permission denied (publickey)"

**Problem**: No SSH key available.

**Solution**:
```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your-email@example.com"

# Try connecting again
ssh -p 4569 -o StrictHostKeyChecking=no localhost
```

### "PTY allocation request failed"

**Problem**: Client not requesting PTY (pseudo-terminal).

**Solution**: The server requires PTY mode. Use standard SSH client:
```bash
# This should work (requests PTY by default)
ssh -p 4569 localhost

# Force PTY allocation if needed
ssh -p 4569 -t localhost
```

### "Rate limit exceeded"

**Problem**: Too many connection attempts from your IP.

**Solution**: Wait 1 minute, then try again. Rate limit resets every minute.

### "Server is at maximum capacity"

**Problem**: 30 users currently connected.

**Solution**: 
- Wait for users to disconnect
- Check active sessions: `make stats`
- Restart server if needed (disconnects all users)

### Port Already in Use

**Problem**: Another process is using port 4569.

**Solution**:
```bash
# Find process using port
sudo lsof -i :4569

# Kill the process
sudo kill -9 <PID>

# Or use different port
make run-server PORT=7022
```

### Binary Not Found

**Problem**: `terminal-web` binary doesn't exist.

**Solution**:
```bash
# Build the binary
make build

# Verify it exists
ls -la terminal-web
```

### Log File Not Created

**Problem**: Logs directory doesn't exist or permissions issue.

**Solution**:
```bash
# Create directories
make setup

# Check permissions
ls -la logs/

# Make directory writable
chmod 755 logs/
```

## Port Forwarding for External Access

If you want to access your local server from the internet:

### 1. Find Your Public IP

```bash
curl ifconfig.me
```

### 2. Configure Router Port Forwarding

1. Access your router admin panel (usually http://192.168.1.1)
2. Find "Port Forwarding" or "Virtual Servers"
3. Add rule:
   - External Port: 4569
   - Internal IP: Your computer's local IP
   - Internal Port: 4569
   - Protocol: TCP

### 3. Test External Access

From outside your network:
```bash
ssh -p 4569 -o StrictHostKeyChecking=no <your-public-ip>
```

**Note**: This exposes your server to the internet. Ensure:
- Firewall is configured
- Rate limiting is active
- You trust the network

## Makefile Commands Reference

| Command | Description |
|---------|-------------|
| `make setup` | Create keys/ and logs/ directories |
| `make gen-key` | Generate SSH host key |
| `make build` | Compile the binary |
| `make run` | Run in local mode (no SSH) |
| `make run-server` | Run SSH server in foreground |
| `make start-server` | Run SSH server in background |
| `make stop-server` | Stop background server |
| `make restart-server` | Restart background server |
| `make status` | Check if server is running |
| `make tail-logs` | View live logs |
| `make stats` | Show connection statistics |
| `make check-security` | Verify security configuration |
| `make clean` | Remove binary |
| `make clean-all` | Remove binary, logs, and keys |
| `make deps` | Install Go dependencies |
| `make fmt` | Format Go code |
| `make lint` | Run linter |
| `make help` | Show all commands |

## Directory Structure

```
terminal-web/
├── terminal-web              # Compiled binary
├── main.go                   # Entry point
├── ssh-server.go            # SSH server implementation
├── config.go                # Configuration
├── logger.go                # Audit logging
├── buble.go                 # TUI logic
├── tui.go                   # TUI parsing
├── lua-util.go              # Lua utilities
├── Makefile                 # Build commands
├── SECURITY.md              # Security documentation
├── LOCAL_SETUP.md           # This file
├── keys/                    # SSH keys (gitignored)
│   ├── ssh_host_ed25519_key
│   ├── ssh_host_ed25519_key.pub
│   └── .gitkeep
├── logs/                    # Audit logs (gitignored)
│   ├── terminal-web.log
│   └── .gitkeep
└── resume/                  # Your resume
    ├── index.html
    └── index.lua
```

## Next Steps

1. **Customize your resume**: Edit `resume/index.html`
2. **Configure firewall**: `sudo ufw allow 4569/tcp`
3. **Point domain**: Create DNS A record to your IP
4. **Share**: Tell people to run `ssh -p 4569 yourdomain.com`

## Support

- Check logs: `tail -f logs/terminal-web.log`
- Verify setup: `make check-security`
- View stats: `make stats`
