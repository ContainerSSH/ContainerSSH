package sshserver

import (
	"errors"
)

// ErrAuthenticationFailed is the error that is returned from TestClient.Connect when the authentication failed.
var ErrAuthenticationFailed = errors.New("authentication failed")

// TestClient is an SSH client intended solely for testing purposes.
type TestClient interface {
	// Connect establishes a connection to the server.
	Connect() (TestClientConnection, error)
	// MustConnect is identical to Connect, but panics it if it cannot connect.
	MustConnect() TestClientConnection
}
