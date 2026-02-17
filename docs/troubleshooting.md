# Troubleshooting

Common issues and their solutions.

## "Connection refused"

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

## "Permission denied (publickey)"

**Problem**: No SSH key available.

**Solution**:
```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your-email@example.com"

# Try connecting again
ssh -p 4569 -o StrictHostKeyChecking=no localhost
```

## "TUI displayed immediately" (Server Mode)

**Problem**: You ran `make run-server` but saw the TUI appear.

**Solution**:
- Make sure you're using: `make run-server` (not `make run`)
- Rebuild after code changes: `make build`
- Stop any existing processes: `make stop-server`

## "Server keeps running after error"

**Problem**: Server process persists after stopping.

**Solution**:
```bash
# Force stop all instances
pkill -9 -f "terminal-web.git -server"

# Verify
make status
```

## "PTY allocation request failed"

**Problem**: Client not requesting PTY (pseudo-terminal).

**Solution**: Use standard SSH client with PTY:
```bash
# This should work (requests PTY by default)
ssh -p 4569 localhost

# Force PTY allocation if needed
ssh -p 4569 -t localhost
```

## "Rate limit exceeded"

**Problem**: Too many connection attempts from your IP.

**Solution**: Wait 1 minute. Rate limit is 10 connections per minute per IP.

## "Server is at maximum capacity"

**Problem**: 30 users currently connected.

**Solution**:
- Wait for users to disconnect
- Check active sessions: `make stats`
- Restart server if needed (disconnects all users)

## Port Already in Use

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

## Firewall Issues

**Ubuntu/Debian:**
```bash
sudo ufw allow 4569/tcp
```

**CentOS/RHEL:**
```bash
sudo firewall-cmd --permanent --add-port=4569/tcp
sudo firewall-cmd --reload
```

**Temporarily disable (testing only):**
```bash
sudo systemctl stop ufw
# or
sudo systemctl stop firewalld
```

## Binary Not Found

**Problem**: `terminal-web` binary doesn't exist.

**Solution**:
```bash
# Build the binary
make build

# Verify it exists
ls -la terminal-web
```

## Log File Not Created

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

## Wrong IP Address

Make sure you're using the server's **local IP address**, not localhost/127.0.0.1:

- ❌ `ssh -p 4569 localhost` (only works on same machine)
- ✅ `ssh -p 4569 192.168.1.100` (works from other computers)

## WSL2 Specific Issues

See the [WSL2 Setup Guide](../setup/wsl2.md#troubleshooting) for WSL2-specific troubleshooting.

## Still Having Issues?

1. Check logs: `tail -f logs/terminal-web.log`
2. Verify setup: `make check-security`
3. View stats: `make stats`
4. Test network connectivity: `ping <server-ip>`
5. Test port: `telnet <server-ip> 4569`
