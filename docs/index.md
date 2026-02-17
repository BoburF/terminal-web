# Terminal-Web Documentation

Welcome to the Terminal-Web SSH server documentation. This project serves your resume as a TUI (Terminal User Interface) application that can be accessed remotely via SSH.

## Quick Navigation

### For Users
- [Getting Started](./getting-started/) - Installation and basic setup
- [Usage Guide](./usage/) - Running and connecting to the server
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions

### For Setup
- [Local Setup](./setup/local.md) - Complete local installation guide
- [WSL2 Setup](./setup/wsl2.md) - Windows WSL2 specific configuration

### For Configuration
- [Security Configuration](./configuration/security.md) - Security features and hardening
- [SSH Protocol Details](./technical/ssh-protocol.md) - Deep dive into SSH implementation

## What is Terminal-Web?

Terminal-Web is a Go application that displays your resume as an interactive terminal-based UI. It can run in two modes:

1. **Local Mode** - Run directly on your terminal for testing
2. **SSH Server Mode** - Allow remote connections via SSH to view your resume

## Features

- **Dual Mode Operation** - Local TUI or SSH server
- **SSH Key Authentication** - Secure public key auth (no passwords)
- **Rate Limiting** - Protects against abuse (10 conn/min per IP)
- **Session Management** - Max 30 concurrent connections, timeouts
- **Audit Logging** - JSON logs of all connection events
- **Customizable Resume** - Edit HTML/Lua files to customize content

## Quick Start

```bash
# Clone and build
git clone <repository-url>
cd terminal-web
make build

# Run locally
make run

# Or start SSH server
make run-server
```

See the [Getting Started Guide](./getting-started/) for detailed instructions.
