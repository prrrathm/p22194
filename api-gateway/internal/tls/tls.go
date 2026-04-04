package tls

import (
	"crypto/tls"
	"fmt"
)

// New builds a *tls.Config from the given certificate and key PEM files.
// It enforces TLS 1.3 as the minimum version and enables modern cipher suites.
func New(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load key pair: %w", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
	return cfg, nil
}
