package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// AuditLogger provides structured JSON logging for security audit trail
type AuditLogger struct {
	file   *os.File
	logger *log.Logger
	mu     sync.Mutex
}

// LogEntry represents a single audit log entry
type LogEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	Level          string    `json:"level"`
	Event          string    `json:"event"`
	IP             string    `json:"ip,omitempty"`
	KeyFingerprint string    `json:"key_fingerprint,omitempty"`
	Duration       string    `json:"duration,omitempty"`
	Message        string    `json:"message"`
	ActiveConns    int       `json:"active_connections,omitempty"`
	Username       string    `json:"username,omitempty"`
	SessionID      string    `json:"session_id,omitempty"`
}

// NewAuditLogger creates a new audit logger that writes to the specified file
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}

	return &AuditLogger{
		file:   file,
		logger: log.New(file, "", 0),
	}, nil
}

// Log writes a structured log entry
func (al *AuditLogger) Log(entry LogEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()

	entry.Timestamp = time.Now().UTC()
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON fails
		al.logger.Printf("ERROR: Failed to marshal log entry: %v", err)
		return
	}
	al.logger.Println(string(data))
}

// LogEvent is a convenience method for logging common events
func (al *AuditLogger) LogEvent(event, ip, message string) {
	al.Log(LogEntry{
		Event:   event,
		IP:      ip,
		Message: message,
	})
}

// LogConnect logs a new connection
func (al *AuditLogger) LogConnect(ip, keyFP, username string) {
	al.Log(LogEntry{
		Level:          "INFO",
		Event:          "CONNECT",
		IP:             ip,
		KeyFingerprint: keyFP,
		Username:       username,
		Message:        "New SSH connection established",
	})
}

// LogDisconnect logs a disconnection with duration
func (al *AuditLogger) LogDisconnect(ip, keyFP string, duration time.Duration) {
	al.Log(LogEntry{
		Level:          "INFO",
		Event:          "DISCONNECT",
		IP:             ip,
		KeyFingerprint: keyFP,
		Duration:       duration.String(),
		Message:        "SSH session ended",
	})
}

// LogRateLimit logs a rate-limited connection attempt
func (al *AuditLogger) LogRateLimit(ip string) {
	al.Log(LogEntry{
		Level:   "WARN",
		Event:   "RATE_LIMIT",
		IP:      ip,
		Message: "Connection rejected due to rate limiting",
	})
}

// LogMaxConnections logs when max connections is reached
func (al *AuditLogger) LogMaxConnections(ip string, currentCount int) {
	al.Log(LogEntry{
		Level:       "WARN",
		Event:       "MAX_CONNECTIONS",
		IP:          ip,
		ActiveConns: currentCount,
		Message:     "Connection rejected: maximum connections reached",
	})
}

// LogAuthFailure logs authentication failures
func (al *AuditLogger) LogAuthFailure(ip, method, reason string) {
	al.Log(LogEntry{
		Level:   "WARN",
		Event:   "AUTH_FAILURE",
		IP:      ip,
		Message: reason,
	})
}

// LogSessionTimeout logs session timeout events
func (al *AuditLogger) LogSessionTimeout(ip, keyFP, reason string) {
	al.Log(LogEntry{
		Level:          "INFO",
		Event:          "SESSION_TIMEOUT",
		IP:             ip,
		KeyFingerprint: keyFP,
		Message:        reason,
	})
}

// LogError logs error events
func (al *AuditLogger) LogError(ip, operation string, err error) {
	al.Log(LogEntry{
		Level:   "ERROR",
		Event:   "ERROR",
		IP:      ip,
		Message: operation + ": " + err.Error(),
	})
}

// Close closes the log file
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()
	return al.file.Close()
}
