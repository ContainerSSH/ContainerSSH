package sshserver

import (
	"time"
)

// TestServer describes
type TestServer interface {
	// GetHostKey returns the hosts private key in PEM format. This can be used to extract the public key.
	GetHostKey() string
	// Start starts the server in the background.
	Start()
	// Stop stops the server running in the background.
	Stop(timeout time.Duration)

	// GetListen returns the listen IP and port
	GetListen() string
}
