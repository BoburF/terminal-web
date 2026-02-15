package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/muesli/termenv"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/net/html"
)

// ConnectionLimiter manages rate limiting and connection counting
type ConnectionLimiter struct {
	connections    map[string][]time.Time // IP -> timestamps of recent connections
	mu             sync.RWMutex
	activeCount    int
	activeMu       sync.RWMutex
	rateLimit      int
	maxConnections int
}

// NewConnectionLimiter creates a new connection limiter
func NewConnectionLimiter(rateLimit, maxConnections int) *ConnectionLimiter {
	return &ConnectionLimiter{
		connections:    make(map[string][]time.Time),
		rateLimit:      rateLimit,
		maxConnections: maxConnections,
	}
}

// allowConnection checks if a connection from the given IP should be allowed
func (cl *ConnectionLimiter) allowConnection(ip string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	minuteAgo := now.Add(-time.Minute)

	// Clean old entries for this IP
	var recent []time.Time
	for _, t := range cl.connections[ip] {
		if t.After(minuteAgo) {
			recent = append(recent, t)
		}
	}
	cl.connections[ip] = recent

	// Check rate limit
	if len(recent) >= cl.rateLimit {
		return false
	}

	// Check max connections
	cl.activeMu.RLock()
	active := cl.activeCount
	cl.activeMu.RUnlock()

	if active >= cl.maxConnections {
		return false
	}

	// Record this connection attempt
	cl.connections[ip] = append(cl.connections[ip], now)
	return true
}

// incrementActive increments the active connection counter
func (cl *ConnectionLimiter) incrementActive() int {
	cl.activeMu.Lock()
	defer cl.activeMu.Unlock()
	cl.activeCount++
	return cl.activeCount
}

// decrementActive decrements the active connection counter
func (cl *ConnectionLimiter) decrementActive() int {
	cl.activeMu.Lock()
	defer cl.activeMu.Unlock()
	if cl.activeCount > 0 {
		cl.activeCount--
	}
	return cl.activeCount
}

// getActiveCount returns current active connection count
func (cl *ConnectionLimiter) getActiveCount() int {
	cl.activeMu.RLock()
	defer cl.activeMu.RUnlock()
	return cl.activeCount
}

// SSHServer represents the secure SSH server
type SSHServer struct {
	config  *Config
	limiter *ConnectionLimiter
	logger  *AuditLogger
	hostKey gossh.Signer
}

// NewSSHServer creates a new SSH server with security configuration
func NewSSHServer(config *Config) (*SSHServer, error) {
	// Initialize audit logger
	logger, err := NewAuditLogger(config.Logging.File)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Load host key
	hostKey, err := loadHostKey(config.Server.HostKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load host key: %w", err)
	}

	return &SSHServer{
		config:  config,
		limiter: NewConnectionLimiter(config.Security.RateLimitPerMinute, config.Security.MaxConnections),
		logger:  logger,
		hostKey: hostKey,
	}, nil
}

// loadHostKey loads the SSH host key from file
func loadHostKey(path string) (gossh.Signer, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key file: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}

	return signer, nil
}

// Start begins listening for SSH connections
func (s *SSHServer) Start() error {
	log.Printf("Starting SSH server on %s:%s...", s.config.Server.Host, s.config.Server.Port)
	log.Printf("Security: max %d connections, %d/min rate limit",
		s.config.Security.MaxConnections, s.config.Security.RateLimitPerMinute)
	log.Printf("Session limits: %v idle timeout, %v max duration",
		s.config.Security.IdleTimeout, s.config.Security.MaxSessionDuration)

	server := &ssh.Server{
		Addr:        net.JoinHostPort(s.config.Server.Host, s.config.Server.Port),
		Handler:     s.secureSessionHandler,
		HostSigners: []ssh.Signer{s.hostKey},

		// Force PTY - prevents non-interactive bots
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		},

		// Accept any public key but log it for audit
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			fingerprint := gossh.FingerprintSHA256(key)
			ip := getClientIP(ctx.RemoteAddr())

			s.logger.Log(LogEntry{
				Level:          "INFO",
				Event:          "AUTH_KEY",
				IP:             ip,
				KeyFingerprint: fingerprint,
				Username:       ctx.User(),
				Message:        "SSH key authentication accepted",
			})

			// Store fingerprint in context for later use
			ctx.SetValue("key_fp", fingerprint)
			return true
		},

		// Reject password authentication
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			ip := getClientIP(ctx.RemoteAddr())
			s.logger.LogAuthFailure(ip, "password", "Password authentication disabled")
			return false
		},

		// Log connection failures
		ConnectionFailedCallback: func(conn net.Conn, err error) {
			ip := getClientIP(conn.RemoteAddr())
			s.logger.LogError(ip, "connection failed", err)
		},
	}

	return server.ListenAndServe()
}

// secureSessionHandler handles individual SSH sessions with security controls
func (s *SSHServer) secureSessionHandler(sess ssh.Session) {
	ip := getClientIP(sess.RemoteAddr())
	keyFP := ""
	if fp, ok := sess.Context().Value("key_fp").(string); ok {
		keyFP = fp
	}
	username := sess.User()
	startTime := time.Now()
	sessionID := generateSessionID(ip, startTime)

	// Check rate limit
	if !s.limiter.allowConnection(ip) {
		if s.limiter.getActiveCount() >= s.config.Security.MaxConnections {
			s.logger.LogMaxConnections(ip, s.limiter.getActiveCount())
			fmt.Fprintln(sess, "Server is at maximum capacity. Please try again later.")
		} else {
			s.logger.LogRateLimit(ip)
			fmt.Fprintln(sess, "Rate limit exceeded. Maximum 10 connections per minute.")
		}
		sess.Exit(1)
		return
	}

	// Increment active connections
	activeCount := s.limiter.incrementActive()
	s.logger.LogConnect(ip, keyFP, username)
	s.logger.Log(LogEntry{
		Level:          "INFO",
		Event:          "SESSION_START",
		IP:             ip,
		KeyFingerprint: keyFP,
		SessionID:      sessionID,
		ActiveConns:    activeCount,
		Message:        fmt.Sprintf("Session started (%d/%d active)", activeCount, s.config.Security.MaxConnections),
	})

	// Decrement active connections on exit
	defer func() {
		duration := time.Since(startTime)
		activeCount := s.limiter.decrementActive()
		s.logger.LogDisconnect(ip, keyFP, duration)
		s.logger.Log(LogEntry{
			Level:          "INFO",
			Event:          "SESSION_END",
			IP:             ip,
			KeyFingerprint: keyFP,
			SessionID:      sessionID,
			Duration:       duration.String(),
			ActiveConns:    activeCount,
			Message:        fmt.Sprintf("Session ended. Duration: %v", duration),
		})
	}()

	// Check PTY requirement
	ptyReq, winCh, isPty := sess.Pty()
	if !isPty {
		s.logger.LogError(ip, "session", fmt.Errorf("no PTY requested"))
		fmt.Fprintln(sess, "PTY is required for this application")
		sess.Exit(1)
		return
	}

	// Open PTY
	ptmx, tty, err := pty.Open()
	if err != nil {
		s.logger.LogError(ip, "pty open", err)
		sess.Exit(1)
		return
	}
	defer ptmx.Close()
	defer tty.Close()

	// Set initial PTY size
	if err := pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(ptyReq.Window.Height),
		Cols: uint16(ptyReq.Window.Width),
	}); err != nil {
		s.logger.LogError(ip, "pty setsize", err)
	}

	// Handle window resize
	go func() {
		for win := range winCh {
			if err := pty.Setsize(ptmx, &pty.Winsize{
				Rows: uint16(win.Height),
				Cols: uint16(win.Width),
			}); err != nil {
				s.logger.LogError(ip, "pty resize", err)
			}
		}
	}()

	// Handle input/output with session timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Security.MaxSessionDuration)
	defer cancel()

	// Copy input from session to PTY
	go func() {
		defer cancel() // Cancel context when input ends
		io.Copy(ptmx, sess)
	}()

	// Copy output from PTY to session
	go func() {
		io.Copy(sess, ptmx)
	}()

	// Run TUI with timeout monitoring
	done := make(chan bool)
	go func() {
		s.runTUIWithTimeout(ctx, tty, ptyReq.Window.Width, ptyReq.Window.Height, sess, sessionID)
		done <- true
	}()

	// Wait for either TUI completion or timeout
	select {
	case <-done:
		// Normal exit
	case <-ctx.Done():
		s.logger.LogSessionTimeout(ip, keyFP, "Maximum session duration reached")
		fmt.Fprintln(sess, "\r\nSession timeout: Maximum duration reached.")
		sess.Exit(0)
	}
}

// runTUIWithTimeout runs the TUI with session timeout handling
func (s *SSHServer) runTUIWithTimeout(ctx context.Context, tty *os.File, width, height int, sess ssh.Session, sessionID string) {
	// Set TERM environment variable for color support in lipgloss (before anything else)
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")

	file, err := os.OpenFile(RootPath+"index.html", os.O_RDONLY, 0o644)
	if err != nil {
		fmt.Fprintf(tty, "Error opening resume: %v\r\n", err)
		return
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		fmt.Fprintf(tty, "Error parsing resume: %v\r\n", err)
		return
	}

	for node := range doc.Descendants() {
		if node.Data == "head" {
			foundScriptToBind(node)
		}

		if node.Data == "body" {
			state, err := drawTui(node)
			if err != nil {
				fmt.Fprintf(tty, "Error creating TUI: %v\r\n", err)
				return
			}
			state.Width = width
			state.Height = height
			state.session = sess

			// Force true color profile for lipgloss rendering
			lipgloss.SetColorProfile(termenv.TrueColor)

			p := tea.NewProgram(
				state,
				tea.WithInput(tty),
				tea.WithOutput(tty),
				tea.WithAltScreen(),
			)

			if _, err := p.Run(); err != nil {
				log.Printf("Error running TUI: %v", err)
			}
		}
	}
}

// getClientIP extracts the client IP from a network address
func getClientIP(addr net.Addr) string {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.IP.String()
	}
	return addr.String()
}

// generateSessionID creates a unique session identifier
func generateSessionID(ip string, startTime time.Time) string {
	return fmt.Sprintf("%s-%d", ip, startTime.UnixNano())
}

// Close gracefully shuts down the server
func (s *SSHServer) Close() error {
	if s.logger != nil {
		return s.logger.Close()
	}
	return nil
}
