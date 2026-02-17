# Usage Guide

Learn how to use Terminal-Web effectively.

## Running Modes

### Local Mode

For testing and development:

```bash
make run
```

The TUI appears directly in your terminal.

### SSH Server - Foreground Mode

Best for testing and debugging:

```bash
make run-server
```

You'll see server logs in real-time. Press `Ctrl+C` to stop.

### SSH Server - Background Mode

For production use:

```bash
# Start server
make start-server

# Check status
make status

# View logs
make tail-logs

# Stop server
make stop-server
```

## Connecting to the Server

### From Local Machine

```bash
ssh -p 4569 -o StrictHostKeyChecking=no localhost
```

### From Another Computer

```bash
ssh -p 4569 -o StrictHostKeyChecking=no <server-ip>
```

**Find your IP address:**
- Linux/macOS: `hostname -I` or `ip addr show`
- Windows: `ipconfig`

### Skip Host Key Verification (First Time)

```bash
ssh -p 4569 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null <server-ip>
```

## Custom Port

```bash
# Server on custom port
make run-server PORT=7022

# Client connects to custom port
ssh -p 7022 <server-ip>
```

## TUI Navigation

Once connected, use these keys:

| Key | Action |
|-----|--------|
| `Tab` | Next section |
| `Shift+Tab` | Previous section |
| `j` or `↓` | Scroll down |
| `k` or `↑` | Scroll up |
| `q` | Exit |
| `Ctrl+C` | Exit |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Compile the binary |
| `make run` | Run locally (TUI mode) |
| `make run-server` | Start SSH server (foreground) |
| `make start-server` | Start SSH server (background) |
| `make stop-server` | Stop background server |
| `make restart-server` | Restart background server |
| `make status` | Check if server is running |
| `make tail-logs` | View live logs |
| `make stats` | Show connection statistics |
| `make check-security` | Verify security configuration |
| `make test-ssh` | Test local SSH connection |
| `make clean` | Remove binary |
| `make clean-all` | Remove binary, logs, and keys |
| `make help` | Show all commands |

## Customizing Your Resume

Edit these files:

- `resume/index.html` - Resume content and structure
- `resume/index.lua` - Key bindings and interactions

After editing, restart the server:

```bash
make restart-server
```

## Log Analysis

View connection logs:

```bash
# View all logs
cat logs/terminal-web.log

# Follow live logs
tail -f logs/terminal-web.log

# View recent connections
jq 'select(.event == "CONNECT")' logs/terminal-web.log | tail -20

# Count unique visitors
jq -r 'select(.event == "CONNECT") | .key_fingerprint' logs/terminal-web.log | sort | uniq | wc -l
```

## External Access

To access from the internet:

1. Find your public IP: `curl ifconfig.me`
2. Configure port forwarding on your router (port 4569 → your computer)
3. Connect using public IP:
   ```bash
   ssh -p 4569 <your-public-ip>
   ```

⚠️ **Warning**: Only expose to the internet with proper security measures. See [Security Configuration](../configuration/security.md).
