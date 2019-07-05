package server

import (
	"errors"

	"github.com/sirupsen/logrus"
)

// Common error definitions for the server
var (
	ErrNoListeners          = errors.New("no listeners")
	ErrIncompleteTLSConfig  = errors.New("incomplete TLS configuration")
	ErrInvalidListenAddress = errors.New("invalid listen address")
)

// ListenerConfig holds the configuration for a listener
type ListenerConfig struct {
	// Address holds the address to lsiten on
	Address string `json:"listen,omitempty"`

	// TLSKeyPath may hold the path to the TLS private key
	TLSKeyPath string `json:"tls_key_path,omitempty"`

	// TLSCertPath may hold the path to the TLS certificate chain
	TLSCertPath string `json:"tls_cert_path,omitempty"`
}

// Config hold server configuration values
type Config struct {
	// Listeners holds all listener configuration
	Listeners []ListenerConfig `json:"listeners,omitempty"`
	Logger    *logrus.Logger
}

// Validate the configuration and returns the first error identified
func (cfg *Config) Validate() error {
	if len(cfg.Listeners) == 0 {
		return ErrNoListeners
	}

	for _, l := range cfg.Listeners {
		if l.Address == "" {
			return ErrInvalidListenAddress
		}

		useTLS := l.TLSKeyPath != "" || l.TLSCertPath != ""
		if useTLS && (l.TLSKeyPath == "" || l.TLSCertPath == "") {
			return ErrIncompleteTLSConfig
		}
	}

	return nil
}
