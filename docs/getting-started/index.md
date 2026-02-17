# Getting Started

This guide will help you get Terminal-Web up and running quickly.

## Prerequisites

- **Go** 1.23 or higher
- **SSH client** (usually pre-installed)
- **Git** (for cloning)

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd terminal-web
```

### 2. Install Dependencies

```bash
make deps
```

This downloads all required Go packages.

### 3. Build the Application

```bash
make build
```

This creates the `terminal-web` binary in the current directory.

### 4. Initial Setup

```bash
make setup      # Create keys/ and logs/ directories
make gen-key    # Generate SSH host key (one-time)
```

## Running the Application

### Local Mode (TUI only)

Run the resume as a local terminal application:

```bash
make run
```

### SSH Server Mode

Start the SSH server to allow remote connections:

```bash
# Foreground mode (see logs in real-time)
make run-server

# Background mode (daemon)
make start-server
```

## Connecting

From another computer on the same network:

```bash
ssh -p 4569 <server-ip-address>
```

Example:
```bash
ssh -p 4569 192.168.1.100
```

## Next Steps

- [Local Setup Guide](../setup/local.md) - Detailed setup instructions
- [Usage Guide](../usage/) - Complete usage documentation
- [Security Configuration](../configuration/security.md) - Security hardening
