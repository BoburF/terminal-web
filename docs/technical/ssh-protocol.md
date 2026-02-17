# SSH (Secure Shell) Protocol Deep Dive

## Table of Contents
1. [Introduction to SSH](#introduction-to-ssh)
2. [SSH Protocol Architecture](#ssh-protocol-architecture)
3. [The SSH Handshake Process](#the-ssh-handshake-process)
4. [Key Exchange (KEX) Algorithms](#key-exchange-kex-algorithms)
5. [Transport Layer Encryption](#transport-layer-encryption)
6. [Authentication Methods](#authentication-methods)
7. [Connection Layer and Channels](#connection-layer-and-channels)
8. [SSH Protocol Details](#ssh-protocol-details)
9. [Implementation Architecture](#implementation-architecture)
10. [Security Considerations](#security-considerations)

---

## Introduction to SSH

SSH (Secure Shell) is a cryptographic network protocol for operating network services securely over an unsecured network. It was designed as a replacement for Telnet and other insecure remote shell protocols.

### Why SSH?

**Before SSH (insecure protocols):**
- **Telnet**: Sent data in plaintext - passwords and commands visible to anyone sniffing the network
- **rsh/rlogin**: Trusted host authentication based on IP addresses - easily spoofed
- **FTP**: Credentials and data transmitted unencrypted

**SSH solves these problems by providing:**
1. **Encryption**: All data is encrypted using symmetric encryption
2. **Authentication**: Strong host and user authentication
3. **Integrity**: Protection against tampering via MAC (Message Authentication Code)
4. **Port Forwarding**: Secure tunneling of other protocols

### SSH Versions

- **SSH-1**: Deprecated due to security vulnerabilities (CRC-32 integrity check is weak)
- **SSH-2**: Current standard (RFC 4251-4254), uses HMAC for integrity

---

## SSH Protocol Architecture

SSH is organized into four conceptual layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    SSH Connection Layer                     │
│     (Channels, Channel Requests, Global Requests)          │
├─────────────────────────────────────────────────────────────┤
│                  SSH Authentication Layer                   │
│        (User Authentication: password, publickey)          │
├─────────────────────────────────────────────────────────────┤
│                   SSH Transport Layer                       │
│     (Key Exchange, Server Authentication, Encryption)      │
├─────────────────────────────────────────────────────────────┤
│                     TCP/IP Layer                            │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

1. **Transport Layer**: Provides server authentication, confidentiality, and integrity
2. **Authentication Layer**: Authenticates the user to the server
3. **Connection Layer**: Multiplexes multiple channels over a single encrypted connection

---

## The SSH Handshake Process

The SSH connection establishment involves multiple phases:

### Phase 1: TCP Connection (3-way handshake)
```
Client                                    Server
   │                                         │
   │ ─────── SYN ─────────────────────────>  │
   │                                         │
   │ <────── SYN-ACK ──────────────────────  │
   │                                         │
   │ ─────── ACK ─────────────────────────>  │
   │                                         │
```

### Phase 2: SSH Protocol Version Exchange

**Client sends:**
```
SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.10
```

**Server responds:**
```
SSH-2.0-GoSSH_0.3.8
```

Both sides MUST send their version string before proceeding.

### Phase 3: Key Exchange Init (KEX_INIT)

Both sides exchange lists of supported algorithms:

**Client sends KEX_INIT:**
```
byte      SSH_MSG_KEXINIT (20)
binary    cookie (16 random bytes)
name-list kex_algorithms                    # e.g., curve25519-sha256, ecdh-sha2-nistp256
name-list server_host_key_algorithms        # e.g., ssh-rsa, ssh-ed25519
name-list encryption_algorithms_client_to_server  # e.g., aes256-gcm@openssh.com
name-list encryption_algorithms_server_to_client
name-list mac_algorithms_client_to_server   # e.g., hmac-sha2-256
name-list mac_algorithms_server_to_client
name-list compression_algorithms_client_to_server  # e.g., none, zlib
name-list compression_algorithms_server_to_client
name-list languages_client_to_server        # (usually empty)
name-list languages_server_to_client
boolean   first_kex_packet_follows
uint32    0 (reserved for future extension)
```

**Algorithm Negotiation Process:**
1. Client and server each send their KEX_INIT with supported algorithms
2. Both independently compute the same negotiated algorithms
3. For each algorithm type, the first algorithm in the client's list that also appears in the server's list is chosen
4. This ensures both sides arrive at the same negotiated algorithm without additional round trips

### Phase 4: Key Exchange (Algorithm-Specific)

**Example: curve25519-sha256**

```
Client                                    Server
   │                                         │
   │ ───── SSH_MSG_KEX_ECDH_INIT ─────────>  │
   │    (client ephemeral public key Q_C)    │
   │                                         │
   │ <──── SSH_MSG_KEX_ECDH_REPLY ─────────  │
   │    (server public key K_S,             │
   │     server ephemeral public key Q_S,   │
   │     exchange hash signature)           │
   │                                         │
```

**The shared secret is computed as:**
```
K = scalar_mult(Q_C, d_S) = scalar_mult(Q_S, d_C)
```

Where `d_C` and `d_S` are the private keys, and `Q_C` and `Q_S` are the public keys.

### Phase 5: New Keys

After key exchange, both sides send:
```
byte SSH_MSG_NEWKEYS (21)
```

This signals that all subsequent packets will be encrypted with the newly negotiated keys.

### Phase 6: Service Request

```
Client ─── SSH_MSG_SERVICE_REQUEST (ssh-userauth) ───> Server
Client <── SSH_MSG_SERVICE_ACCEPT ─────────────────── Server
```

---

## Key Exchange (KEX) Algorithms

### Available KEX Algorithms

| Algorithm | Security Level | Performance | Notes |
|-----------|---------------|-------------|-------|
| `curve25519-sha256` | High | Fast | Modern, recommended |
| `ecdh-sha2-nistp256` | High | Fast | NIST curve, widely supported |
| `ecdh-sha2-nistp384` | Very High | Medium | Stronger than P256 |
| `ecdh-sha2-nistp521` | Very High | Slower | Strongest NIST curve |
| `diffie-hellman-group-exchange-sha256` | High | Slow | Traditional DH |
| `diffie-hellman-group14-sha256` | Medium | Slow | Legacy, avoid if possible |

### How KEX Generates Keys

**Key Derivation Process:**

```
# Exchange hash (H) is computed over:
# - Client version string
# - Server version string
# - Client KEX_INIT
# - Server KEX_INIT
# - Server host key (K_S)
# - Client ephemeral public key (Q_C)
# - Server ephemeral public key (Q_S)
# - Shared secret (K)

# Initial IV client-to-server: HASH(K || H || "A" || session_id)
# Initial IV server-to-client: HASH(K || H || "B" || session_id)
# Encryption key client-to-server: HASH(K || H || "C" || session_id)
# Encryption key server-to-client: HASH(K || H || "D" || session_id)
# Integrity key client-to-server: HASH(K || H || "E" || session_id)
# Integrity key server-to-client: HASH(K || H || "F" || session_id)
```

---

## Transport Layer Encryption

### Cipher Algorithms

| Cipher | Mode | Key Size | Security |
|--------|------|----------|----------|
| `aes256-gcm@openssh.com` | GCM | 256-bit | Authenticated encryption |
| `aes128-gcm@openssh.com` | GCM | 128-bit | Authenticated encryption |
| `aes256-ctr` | CTR | 256-bit | Good performance |
| `aes192-ctr` | CTR | 192-bit | Good performance |
| `aes128-ctr` | CTR | 128-bit | Good performance |
| `chacha20-poly1305@openssh.com` | Stream | 256-bit | Modern, fast |

### MAC (Message Authentication Code) Algorithms

Used when cipher doesn't provide authenticated encryption:

| MAC Algorithm | Tag Size | Notes |
|--------------|----------|-------|
| `hmac-sha2-256` | 32 bytes | Recommended |
| `hmac-sha2-512` | 64 bytes | Stronger |
| `hmac-sha1` | 20 bytes | Legacy, acceptable |

### SSH Binary Packet Protocol

**Packet Structure:**
```
uint32    packet_length          # Length of packet (excluding length field and padding)
byte      padding_length         # Length of padding
byte[n1]  payload                # Actual data (n1 = packet_length - padding_length - 1)
byte[n2]  random padding         # n2 = padding_length
byte[m]   MAC (if not AEAD)      # Message authentication code
```

**Encryption Process:**
1. Serialize payload into binary format
2. Calculate required padding (minimum 4 bytes, block size alignment)
3. Concatenate: padding_length + payload + padding
4. Encrypt with negotiated cipher
5. Append MAC (if required)

**Decryption Process:**
1. Read packet_length (first 4 bytes, may be encrypted)
2. Read remaining encrypted data
3. Decrypt using negotiated cipher
4. Verify MAC (if required)
5. Extract and parse payload

---

## Authentication Methods

### 1. Password Authentication

**Flow:**
```
Client ─── SSH_MSG_USERAUTH_REQUEST (password) ───> Server
                    │
                    ▼
            [Server validates]
                    │
Client <──────── SSH_MSG_USERAUTH_SUCCESS ──────── Server
         OR
Client <──────── SSH_MSG_USERAUTH_FAILURE ──────── Server
```

**Request format:**
```
byte      SSH_MSG_USERAUTH_REQUEST (50)
string    user name
string    service name (e.g., "ssh-connection")
string    "password"
boolean   FALSE (not changing password)
string    plaintext password in ISO-10646 UTF-8 encoding
```

### 2. Public Key Authentication

**Two-phase process:**

**Phase 1 - Probe:**
```
Client ─── SSH_MSG_USERAUTH_REQUEST (publickey, FALSE) ───> Server
Client <──────────────── SSH_MSG_USERAUTH_PK_OK ───────── Server
```

**Phase 2 - Authenticate:**
```
Client ─── SSH_MSG_USERAUTH_REQUEST (publickey, TRUE, signature) ───> Server
Client <────────────────── SSH_MSG_USERAUTH_SUCCESS ─────────────── Server
```

**Signature computation:**
```
sign( H ||
      string  session identifier (H from first KEX)
      byte    SSH_MSG_USERAUTH_REQUEST
      string  user name
      string  service
      string  "publickey"
      boolean TRUE
      string  public key algorithm name
      string  public key blob
)
```

### 3. Keyboard-Interactive Authentication

Used for multi-factor authentication, PAM, etc.

**Flow:**
```
Client ─── SSH_MSG_USERAUTH_REQUEST (keyboard-interactive) ───> Server
Client <──────── SSH_MSG_USERAUTH_INFO_REQUEST ─────────────── Server
                  (name, instruction, language, [prompts...])
Client ─── SSH_MSG_USERAUTH_INFO_RESPONSE ───────────────────> Server
                  ([responses...])
Client <──────────── SSH_MSG_USERAUTH_SUCCESS ──────────────── Server
```

---

## Connection Layer and Channels

### Channel Types

| Channel Type | Purpose |
|--------------|---------|
| `session` | Remote execution of a program (shell, command, subsystem) |
| `direct-tcpip` | Client-to-server port forwarding |
| `forwarded-tcpip` | Server-to-client port forwarding (remote forwarding) |
| `x11` | X11 forwarding |
| `auth-agent@openssh.com` | SSH agent forwarding |

### Channel Lifecycle

**Opening a Channel:**
```
Client ─── SSH_MSG_CHANNEL_OPEN ─────────────────────> Server
           (channel type, sender channel, 
            initial window size, max packet size)

Client <── SSH_MSG_CHANNEL_OPEN_CONFIRMATION ──────── Server
           (recipient channel, sender channel,
            initial window size, max packet size)
```

**Data Transfer:**
```
Client ─── SSH_MSG_CHANNEL_DATA ─────────────────────> Server
           (recipient channel, data)

Client <── SSH_MSG_CHANNEL_DATA ───────────────────── Server
           (recipient channel, data)
```

**Closing a Channel:**
```
Client ─── SSH_MSG_CHANNEL_EOF ──────────────────────> Server
Client ─── SSH_MSG_CHANNEL_CLOSE ────────────────────> Server
Client <── SSH_MSG_CHANNEL_CLOSE ──────────────────── Server
```

### Flow Control (Windowing)

SSH uses a credit-based flow control:

```
Initial Window Size: 2097152 bytes (2MB default)

As data is sent:
  Window decreases by data size

When window is low:
  Receiver sends SSH_MSG_CHANNEL_WINDOW_ADJUST
  to add more credits

This prevents overwhelming the receiver
```

### Channel Requests

Requests are sent over established channels:

| Request Type | Direction | Purpose |
|-------------|-----------|---------|
| `pty-req` | Client → Server | Allocate pseudo-terminal |
| `x11-req` | Client → Server | Request X11 forwarding |
| `env` | Client → Server | Set environment variable |
| `shell` | Client → Server | Start shell |
| `exec` | Client → Server | Execute command |
| `subsystem` | Client → Server | Start subsystem (e.g., sftp) |
| `window-change` | Client → Server | Terminal window size changed |
| `signal` | Client → Server | Send signal to process |
| `exit-status` | Server → Client | Command exit status |
| `exit-signal` | Server → Client | Process terminated by signal |

---

## SSH Protocol Details

### Message Types (Decimal Values)

| Value | Message | Description |
|-------|---------|-------------|
| 1-19 | Transport layer generic |
| 20 | SSH_MSG_KEXINIT | Key exchange init |
| 21 | SSH_MSG_NEWKEYS | New keys acknowledged |
| 30-49 | Key exchange method specific |
| 50-59 | User authentication |
| 50 | SSH_MSG_USERAUTH_REQUEST | Authentication request |
| 51 | SSH_MSG_USERAUTH_FAILURE | Authentication failed |
| 52 | SSH_MSG_USERAUTH_SUCCESS | Authentication succeeded |
| 60-79 | Connection protocol |
| 80-89 | Reserved for client protocols |
| 90-127 | Channel related messages |
| 90 | SSH_MSG_CHANNEL_OPEN | Open new channel |
| 91 | SSH_MSG_CHANNEL_OPEN_CONFIRMATION | Channel open success |
| 92 | SSH_MSG_CHANNEL_OPEN_FAILURE | Channel open failed |
| 93 | SSH_MSG_CHANNEL_WINDOW_ADJUST | Adjust window size |
| 94 | SSH_MSG_CHANNEL_DATA | Channel data |
| 95 | SSH_MSG_CHANNEL_EXTENDED_DATA | Extended data (stderr) |
| 96 | SSH_MSG_CHANNEL_EOF | EOF on channel |
| 97 | SSH_MSG_CHANNEL_CLOSE | Close channel |
| 98 | SSH_MSG_CHANNEL_REQUEST | Channel request |
| 100 | SSH_MSG_CHANNEL_SUCCESS | Request succeeded |
| 101 | SSH_MSG_CHANNEL_FAILURE | Request failed |

### Data Types

| Type | Description |
|------|-------------|
| `byte` | 8-bit byte |
| `boolean` | TRUE(1) or FALSE(0) |
| `uint32` | 32-bit unsigned integer (MSB first) |
| `uint64` | 64-bit unsigned integer (MSB first) |
| `string` | uint32 length + byte[] (no null termination) |
| `mpint` | Multiple precision integer (signed) |
| `name-list` | Comma-separated list of names |

### Example: Encoding a String

```
String: "ssh-userauth"

Encoding:
00 00 00 0c 73 73 68 2d 75 73 65 72 61 75 74 68
│  │  │  │  └───┬──┴───┬──┴───┬──┴───┬──┴───┘
│  │  │  │      s      s      h      -
│  │  └──┴── length = 12 (0x0000000c)
└──┴── length bytes (MSB)
```

---

## Implementation Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                        SSH Server                              │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Listener   │  │  Connection  │  │   Session Manager    │  │
│  │   (TCP)      │  │   Handler    │  │                      │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                 │                      │              │
│         └─────────────────┼──────────────────────┘              │
│                           ▼                                     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    SSH Protocol Stack                     │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  ┌───────────┐  ┌───────────┐  ┌─────────────────────┐  │  │
│  │  │ Connection│  │    Auth   │  │      Transport      │  │  │
│  │  │  Layer    │  │   Layer   │  │       Layer         │  │  │
│  │  └─────┬─────┘  └─────┬─────┘  └──────────┬──────────┘  │  │
│  │        └──────────────┼───────────────────┘             │  │
│  │                       ▼                                 │  │
│  │  ┌───────────────────────────────────────────────────┐  │  │
│  │  │              Packet Encoder/Decoder                │  │  │
│  │  └───────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           │                                     │
│  ┌────────────────────────┼────────────────────────────────┐   │
│  │                        ▼                                │   │
│  │  ┌───────────┐  ┌───────────┐  ┌─────────────────────┐  │   │
│  │  │  KEX      │  │  Cipher   │  │       MAC          │  │   │
│  │  │ Handler   │  │   (AES)   │  │    (HMAC-SHA2)     │  │   │
│  │  └───────────┘  └───────────┘  └─────────────────────┘  │   │
│  └───────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Transport Layer**
   - Handles TCP connection
   - Manages version exchange
   - Performs key exchange
   - Encrypts/decrypts packets
   - Verifies integrity

2. **Authentication Layer**
   - Validates user credentials
   - Supports multiple auth methods
   - Handles partial success (multi-factor)

3. **Connection Layer**
   - Manages channels
   - Handles requests
   - Implements flow control
   - Routes data

4. **Session Management**
   - Tracks active sessions
   - Manages PTY allocation
   - Handles process execution

---

## Security Considerations

### Host Key Verification

**Why it matters:**
- Prevents man-in-the-middle attacks
- Ensures you're connecting to the right server

**First connection:**
```
The authenticity of host 'example.com (192.168.1.100)' can't be established.
ED25519 key fingerprint is SHA256:abc123...
Are you sure you want to continue connecting? (yes/no)
```

**Known hosts file (~/.ssh/known_hosts):**
```
example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...
192.168.1.100 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...
```

### Best Practices

1. **Disable password authentication** - Use public keys only
2. **Use strong algorithms** - Prefer ed25519, curve25519-sha256
3. **Limit user access** - Use AllowUsers, DenyUsers
4. **Keep software updated** - Patch vulnerabilities
5. **Use fail2ban** - Prevent brute force attacks
6. **Monitor logs** - Detect suspicious activity
7. **Disable root login** - Use unprivileged users
8. **Use key passphrases** - Protect private keys

### Algorithm Recommendations

**Recommended (Modern):**
- KEX: `curve25519-sha256`, `ecdh-sha2-nistp521`
- Host Key: `ssh-ed25519`, `ecdsa-sha2-nistp521`
- Cipher: `chacha20-poly1305@openssh.com`, `aes256-gcm@openssh.com`
- MAC: `hmac-sha2-256-etm@openssh.com` (encrypt-then-mac)

**Acceptable (Widely supported):**
- KEX: `diffie-hellman-group-exchange-sha256`
- Host Key: `rsa-sha2-512`
- Cipher: `aes256-ctr`

**Avoid (Weak/Deprecated):**
- KEX: `diffie-hellman-group1-sha1`, `diffie-hellman-group14-sha1`
- Host Key: `ssh-dss`, `ssh-rsa` (using SHA-1)
- Cipher: `3des-cbc`, `blowfish-cbc`, `arcfour`
- MAC: `hmac-md5`, `hmac-sha1` (without -etm)

---

## Summary

SSH is a sophisticated protocol with multiple layers working together to provide secure remote access:

1. **Transport Layer** establishes encrypted channel through key exchange
2. **Authentication Layer** verifies user identity securely
3. **Connection Layer** multiplexes multiple channels over one connection
4. **Application Layer** (your code) uses channels for shell, commands, forwarding

Understanding these layers helps in building secure, efficient SSH servers and clients.

---

## References

- RFC 4251 - The Secure Shell (SSH) Protocol Architecture
- RFC 4252 - SSH Authentication Protocol
- RFC 4253 - SSH Transport Layer Protocol
- RFC 4254 - SSH Connection Protocol
- RFC 5656 - ECDH Key Exchange
- RFC 6668 - SHA-2 MAC algorithms
- RFC 8308 - Extension Negotiation
