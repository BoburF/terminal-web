# Project Architecture

Overview of the Terminal-Web project architecture and code organization.

## Project Structure

```
terminal-web/
├── main.go              # Application entry point
├── ssh-server.go        # SSH server implementation
├── config.go            # Configuration management
├── logger.go            # Audit logging system
├── buble.go             # TUI (Bubble Tea) logic
├── tui.go               # TUI rendering and parsing
├── lua-util.go          # Lua scripting utilities
├── Makefile             # Build automation
├── go.mod               # Go module definition
├── go.sum               # Go dependency checksums
├── docs/                # Documentation
├── keys/                # SSH host keys (gitignored)
├── logs/                # Audit logs (gitignored)
├── resume/              # Resume content
│   ├── index.html       # Resume HTML content
│   └── index.lua        # Resume interactions
└── examples/            # Example configurations
```

## Core Components

### 1. Entry Point (`main.go`)

Handles command-line arguments and mode selection:
- Local TUI mode (default)
- SSH server mode (`-server` flag)
- Custom port override (`-port` flag)

### 2. SSH Server (`ssh-server.go`)

Implements the SSH server with security features:
- TCP listener
- SSH handshake and key exchange
- Public key authentication
- Rate limiting
- Connection limiting
- Session management
- PTY handling
- TUI integration

### 3. Configuration (`config.go`)

Centralized configuration management:
- Server settings (port, timeouts)
- Security limits (rate limits, connection limits)
- Host key paths
- Log file paths

### 4. Audit Logger (`logger.go`)

JSON-based audit logging:
- Connection events
- Authentication events
- Rate limiting events
- Session lifecycle events
- Error tracking

### 5. TUI System (`buble.go`, `tui.go`)

Terminal User Interface implementation:
- HTML parsing for resume content
- Bubble Tea framework integration
- Keyboard event handling
- View rendering
- Window resizing

### 6. Lua Integration (`lua-util.go`)

Scripting support for dynamic content:
- Lua VM initialization
- Function bindings
- Resume interaction logic

## Data Flow

### Local Mode

```
main.go → runLocalMode() → HTML Parse → TUI State → Bubble Tea Program
```

### SSH Server Mode

```
main.go → SSH Server Start → TCP Listen → Connection Handler
                                           ↓
                        Key Exchange → Authentication → PTY Allocation
                                           ↓
                                 TUI Program Start
```

## Security Architecture

### Authentication Flow

1. **TCP Connection** - Client connects to port 4569
2. **Version Exchange** - SSH protocol version negotiation
3. **Key Exchange** - Diffie-Hellman key exchange
4. **Service Request** - Request ssh-userauth service
5. **Public Key Auth** - Validate client's SSH key
6. **Channel Open** - Open session channel
7. **PTY Request** - Allocate pseudo-terminal
8. **Shell/Exec** - Start TUI program

### Rate Limiting

- Per-IP tracking using sliding window
- 10 connections per minute limit
- Automatic cleanup of old records

### Connection Limiting

- Global limit of 30 concurrent connections
- Graceful rejection with user-friendly message
- Connection tracking with unique session IDs

### Session Management

- Idle timeout: 5 minutes
- Maximum session duration: 10 minutes
- Automatic cleanup on timeout
- Graceful disconnection

## Dependencies

### Core Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `golang.org/x/crypto/ssh` - SSH server implementation
- `golang.org/x/net/html` - HTML parsing
- `github.com/yuin/gopher-lua` - Lua scripting

### Build Dependencies

- Go 1.23+
- Make
- OpenSSH client (for testing)

## Development Workflow

### Building

```bash
make build          # Compile binary
make clean          # Remove binary
make deps           # Install dependencies
```

### Testing

```bash
make run            # Local TUI mode
make run-server     # SSH server foreground
make test-ssh       # Test SSH connection
```

### Code Quality

```bash
make fmt            # Format Go code
make lint           # Run linter
```

## Extension Points

### Adding New Authentication Methods

Modify `ssh-server.go`:
- Add handler in `authHandler`
- Update authentication callback

### Customizing TUI

Modify:
- `resume/index.html` - Content structure
- `resume/index.lua` - Interaction logic
- `tui.go` - Rendering logic

### Adding New Commands

Modify `Makefile`:
- Add new targets for custom operations

## Performance Considerations

### Memory Management

- Connection pooling for active sessions
- Automatic cleanup of disconnected clients
- Log rotation for long-running servers

### CPU Usage

- Efficient HTML parsing
- Minimal allocations in hot paths
- Connection rate limiting to prevent abuse

### Network Optimization

- PTY mode only (no shell access)
- Optimized TUI rendering
- Connection timeouts
