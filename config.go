package main

import (
	"time"
)

// Config holds all server configuration
type Config struct {
	Server   ServerConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Host    string
	Port    string
	HostKey string
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	MaxConnections     int
	RateLimitPerMinute int
	IdleTimeout        time.Duration
	MaxSessionDuration time.Duration
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string
	File  string
}

// DefaultConfig returns the default configuration matching user requirements:
// - Port: 4569
// - Max Connections: 30
// - Rate Limit: 10/minute per IP
// - Idle Timeout: 5 minutes
// - Max Session: 10 minutes
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    "22",
			HostKey: "keys/ssh_host_ed25519_key",
		},
		Security: SecurityConfig{
			MaxConnections:     30,
			RateLimitPerMinute: 10,
			IdleTimeout:        5 * time.Minute,
			MaxSessionDuration: 10 * time.Minute,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "logs/terminal-web.log",
		},
	}
}
