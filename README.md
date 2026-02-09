# Terminal-Web SSH Server Documentation

## Overview
This application can run as both a local TUI (Terminal User Interface) application and as an SSH server that allows remote connections to view the resume.

## Prerequisites
- Go 1.25.4 or higher
- SSH client on the connecting computer
- Both computers must be on the same network (or port forwarding configured)

## Installation

1. Clone or download the repository
2. Install dependencies:
```bash
make deps
```

3. Build the application:
```bash
make build
```

## Running the SSH Server

### Option 1: Run in Foreground (Terminal Mode) - RECOMMENDED
Use this when you want to see server logs and easily stop with Ctrl+C:

```bash
# Run SSH server in current terminal (shows live logs)
make run-server

# The server runs in your terminal - you'll see:
# - "Starting SSH server on port 4444"
# - Connection logs when clients connect
# - Press Ctrl+C to stop
```

### Option 2: Run in Background (Daemon Mode)
Use this when you want the server to keep running after you close the terminal:

```bash
# Start server in background
make start-server

# Check if it's running
make status

# View logs
tail -f /tmp/terminal-web.log

# Stop the background server
make stop-server
```

### Custom Port
```bash
# Foreground with custom port
make run-server PORT=3333

# Background with custom port
make start-server PORT=3333
```

### Get Your IP Address
On the server computer, find your IP address:

**Linux/macOS:**
```bash
hostname -I
# or
ip addr show
```

**Windows:**
```cmd
ipconfig
```

Look for your local IP (usually starts with 192.168.x.x or 10.x.x.x)

### Stop the Server
```bash
make stop-server
```

## Connecting from Another Computer

### Basic Connection
From the client computer, run:

```bash
ssh -p 2222 <server-ip>
```

**Example:**
```bash
ssh -p 2222 192.168.1.100
```

### Skip Host Key Verification (First Time)
```bash
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null <server-ip>
```

### Specify a User (Optional)
Since authentication is disabled, any username works:
```bash
ssh -p 2222 -o StrictHostKeyChecking=no user@192.168.1.100
```

## Controls in the TUI

Once connected, you can use these keys:

- **Ctrl+C** or **q** - Exit the application
- **Tab** - Switch sections (if configured)

## Troubleshooting

### "TUI displayed immediately" or "Error after quitting"
**Problem:** You ran `make run-server` but saw the TUI appear immediately instead of server logs.

**Solution:** 
- This shouldn't happen with `-server` flag. Make sure you're using:
  ```bash
  make run-server    # Correct - starts SSH server
  ```
  NOT:
  ```bash
  make run           # Wrong - this starts local TUI mode!
  ```

- If you see TUI immediately, check:
  1. Did you rebuild after code changes? Run: `make build`
  2. Is there another process running? Run: `make stop-server` then try again
  3. Check you're in the correct directory with the binary

### "Server keeps running after error"
**Problem:** You stopped the server but it's still running in background.

**Solution:**
```bash
# Force stop all instances
pkill -9 -f "terminal-web.git -server"

# Then verify
make status
```

### Connection Refused
1. Check if the server is running:
   ```bash
   make status
   # or
   ps aux | grep terminal-web
   ```

2. Verify the port is listening:
   ```bash
   netstat -tlnp | grep 4444
   ```

### Firewall Issues

**Ubuntu/Debian:**
```bash
sudo ufw allow 2222/tcp
```

**CentOS/RHEL:**
```bash
sudo firewall-cmd --permanent --add-port=2222/tcp
sudo firewall-cmd --reload
```

**Temporarily disable firewall (for testing only):**
```bash
sudo systemctl stop ufw
# or
sudo systemctl stop firewalld
```

### Wrong IP Address
Make sure you're using the server's **local IP address**, not localhost/127.0.0.1:
- ❌ `ssh -p 2222 localhost` (only works on same machine)
- ✅ `ssh -p 2222 192.168.1.100` (works from other computers)

### Port Already in Use
If port 2222 is taken, use a different port:
```bash
# Server
make run-server PORT=3333

# Client
ssh -p 3333 192.168.1.100
```

## Network Requirements

### Local Network (Same WiFi/Router)
- Works automatically
- Use the server's local IP address

### Different Networks (Internet)
You need to:
1. Find your public IP: `curl ifconfig.me`
2. Configure port forwarding on your router:
   - Forward external port 2222 to internal IP:2222
3. Connect using public IP:
   ```bash
   ssh -p 2222 <your-public-ip>
   ```

**Note:** Public internet access requires proper security measures (authentication, firewall rules, etc.)

## Quick Start Example

**On Server (192.168.1.100):**
```bash
cd /path/to/terminal-web

# Run in foreground (recommended for testing)
make run-server
# You'll see: "Starting SSH server on port 4444..."
# The server waits for connections - NO TUI appears here!

# OR run in background
make start-server
```

**On Client:**
```bash
ssh -p 4444 -o StrictHostKeyChecking=no 192.168.1.100
```

**You'll see the resume TUI on the CLIENT!** Press **q** or **Ctrl+C** to exit.

**Important:** The TUI only appears on the client side when someone connects via SSH. The server just shows logs.

## Makefile Commands Reference

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make run` | Run locally (current terminal) |
| `make run-server` | Start SSH server (foreground) |
| `make start-server` | Start SSH server (background) |
| `make stop-server` | Stop background SSH server |
| `make test-ssh` | Test local SSH connection |
| `make clean` | Remove built binary |
| `make help` | Show all available commands |

## Customizing the Resume

Edit the files in the `resume/` directory:
- `index.html` - Resume content and structure
- `index.lua` - Key bindings and interactions

After editing, restart the server:
```bash
make stop-server
make start-server
```

## Security Notes

⚠️ **Warning:** By default, the SSH server accepts any connection without authentication. This is suitable for:
- Local network demos
- Trusted environments
- Quick testing

**Do not expose to the public internet without:**
- Adding authentication (password or SSH keys)
- Using a firewall to restrict access
- Running behind a reverse proxy

## Support

For issues or questions, check:
1. Server logs (displayed in terminal)
2. Firewall settings
3. Network connectivity with `ping <server-ip>`
4. Port availability with `telnet <server-ip> 2222`
