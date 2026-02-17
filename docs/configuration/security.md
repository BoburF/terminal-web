# Security Configuration

This document describes the security features implemented for the public SSH resume server.

## Configuration

| Setting | Value |
|---------|-------|
| **Port** | 4569 |
| **Authentication** | SSH Public Key Only (no passwords) |
| **Max Concurrent Connections** | 30 |
| **Rate Limit** | 10 connections/minute per IP |
| **Idle Timeout** | 5 minutes |
| **Max Session Duration** | 10 minutes |
| **Host Key Algorithm** | Ed25519 |
| **Startup** | Manual |

## Security Features

### 1. SSH Key Authentication
- **No password authentication** - Only SSH public keys are accepted
- **Audit logging** - Every connection is logged with key fingerprint
- **Anonymous-friendly** - Any valid SSH key is accepted, enabling public access without registration

### 2. Rate Limiting
- **Per-IP limiting** - Maximum 10 connection attempts per minute from a single IP
- **Automatic cleanup** - Old connection records are automatically purged
- **Prevents abuse** - Stops brute force and automated scanning

### 3. Connection Management
- **Global limit** - Maximum 30 concurrent connections across all users
- **Session tracking** - Each session has a unique ID and is monitored
- **Graceful handling** - Users see a friendly message when server is at capacity

### 4. Session Timeouts
- **Idle timeout** - Sessions are terminated after 5 minutes of inactivity
- **Max duration** - Hard limit of 10 minutes per session
- **Resource protection** - Prevents resource exhaustion from abandoned sessions

### 5. Forced PTY Mode
- **Interactive only** - All sessions must request a pseudo-terminal
- **Bot prevention** - Automated tools and scripts are blocked
- **TUI enforcement** - Users can only interact with the resume interface

### 6. Audit Logging
All events are logged to `logs/terminal-web.log` in JSON format:

```json
{"timestamp":"2026-02-15T10:30:00Z","event":"CONNECT","ip":"203.0.113.42","key_fingerprint":"SHA256:abc123...","username":"user","message":"New SSH connection established"}
```

**Logged Events:**
- `CONNECT` - New SSH connection established
- `DISCONNECT` - Session ended with duration
- `AUTH_KEY` - SSH key authentication accepted
- `AUTH_FAILURE` - Authentication rejected
- `RATE_LIMIT` - Connection rejected due to rate limiting
- `MAX_CONNECTIONS` - Connection rejected due to capacity
- `SESSION_TIMEOUT` - Session terminated due to timeout
- `ERROR` - Error events with details

## File Structure

```
terminal-web/
├── keys/
│   ├── ssh_host_ed25519_key          # Private host key (600 permissions)
│   ├── ssh_host_ed25519_key.pub      # Public host key
│   └── .gitkeep
├── logs/
│   ├── terminal-web.log              # JSON audit logs
│   └── .gitkeep
├── resume/
│   ├── index.html                    # Your resume
│   └── index.lua
├── terminal-web                      # Binary
└── [source files]
```

## Commands

### Setup
```bash
# Generate host key (one-time setup)
make gen-key

# Verify configuration
make check-security

# Configure firewall (Ubuntu/Debian)
sudo ufw allow 4569/tcp
```

### Run Server
```bash
# Run in foreground (recommended for testing)
make run-server

# Run in background
make start-server

# Stop server
make stop-server

# Check status
make status
```

### Monitoring
```bash
# View live logs
tail -f logs/terminal-web.log

# Or use make command
make tail-logs

# Show connection statistics
make stats
```

### Testing
```bash
# Test from same machine
ssh -p 4569 -o StrictHostKeyChecking=no localhost

# Test from another machine
ssh -p 4569 -o StrictHostKeyChecking=no <your-server-ip>
```

## For Visitors

### First Time Setup
Visitors need to generate an SSH key if they don't have one:

```bash
ssh-keygen -t ed25519 -C "your-email@example.com"
```

### Connecting
```bash
ssh -p 4569 yourdomain.com
```

### Navigation
Once connected:
- **Tab** - Switch to next section
- **Shift+Tab** - Switch to previous section  
- **j** or `↓` - Scroll down
- **k** or `↑` - Scroll up
- **q** or **Ctrl+C** - Exit

## Security Checklist

Before going public:

- [ ] Host key generated (`make gen-key`)
- [ ] Host key permissions correct (600)
- [ ] Firewall configured (port 4569 open)
- [ ] Log directory exists
- [ ] Resume files in place
- [ ] Binary compiled successfully
- [ ] Tested locally
- [ ] Tested from another device

## Troubleshooting

### "Connection refused"
```bash
# Check if server is running
make status

# Check firewall
sudo ufw status | grep 4569
```

### "Permission denied (publickey)"
Users need to generate an SSH key first:
```bash
ssh-keygen -t ed25519
```

### "Rate limit exceeded"
Wait 1 minute between connection attempts. The rate limit is 10 connections per minute per IP.

### "Server is at maximum capacity"
The server allows 30 concurrent users. Wait for someone to disconnect.

## Log Analysis

### View recent connections
```bash
jq 'select(.event == "CONNECT")' logs/terminal-web.log | tail -20
```

### Count unique visitors (by key fingerprint)
```bash
jq -r 'select(.event == "CONNECT") | .key_fingerprint' logs/terminal-web.log | sort | uniq | wc -l
```

### Check for abuse attempts
```bash
jq 'select(.event == "RATE_LIMIT" or .event == "AUTH_FAILURE")' logs/terminal-web.log
```

### Average session duration
```bash
jq -r 'select(.event == "DISCONNECT") | .duration' logs/terminal-web.log | \
  awk -F'm|s' '{sum+=$1*60+$2; count++} END {print "Average: " sum/count " seconds"}'
```

## Domain Setup

When you're ready to point your domain:

1. **DNS Configuration:**
   ```
   Type: A
   Name: resume (or @ for root)
   Value: <your-server-public-ip>
   TTL: 300
   ```

2. **Users connect with:**
   ```bash
   ssh -p 4569 resume.yourdomain.com
   ```

## Security Notes

- **Host key fingerprint** - First-time users will see a prompt asking to verify the host key
- **No shell access** - Users can only view the TUI, cannot execute commands
- **Read-only** - The resume is read-only, no file system access
- **Encrypted** - All traffic is encrypted using SSH protocol
- **Anonymous** - No personal data is collected, only key fingerprints for audit

## Contact

For issues or questions, check the logs at `logs/terminal-web.log`.
