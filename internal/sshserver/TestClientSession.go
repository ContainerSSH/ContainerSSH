package sshserver

import (
	"context"
	"io"
)

// TestClientSession is a representation of a session channel inside a test client connection.
type TestClientSession interface {
	// SetEnv sets an environment variable or returns with an error.
	SetEnv(name string, value string) error

	// MustSetEnv is identical to SetEnv, but panics if an error happens.
	MustSetEnv(name string, value string)

	// Window requests the terminal window to be resized to a certain size.
	Window(cols int, rows int) error

	// MustWindow is identical to Window, but panics if an error happens.
	MustWindow(cols int, rows int)

	// RequestPTY requests the server to open a PTY/TTY for this channel. Returns an error if the request failed.
	RequestPTY(term string, cols int, rows int) error

	// MustRequestPTY is identical to RequestPTY but panics if an error happens.
	MustRequestPTY(term string, cols int, rows int)

	// Signal sends a signal to the process
	Signal(signal string) error

	// MustSignal is equal to Signal but panics if an error happens.
	MustSignal(signal string)

	// Shell requests a shell to be opened. After this call returns I/O interactions are possible.
	Shell() error

	// MustShell is identical to Shell but panics if an error happens.
	MustShell()

	// Exec requests a specific program to be executed. After this call returns I/O interactions are possible.
	Exec(program string) error

	// MustExec is identical to Exec but panics if an error happens.
	MustExec(program string)

	// Subsystem requests a specific subsystem to be executed. After this call returns I/O interactions are possible.
	Subsystem(name string) error

	// MustSubsystem is identical to Subsystem but panics if an error happens.
	MustSubsystem(name string)

	// Write writes to the stdin of the session.
	Write(data []byte) (int, error)

	// Type writes to the stdin slowly with 50 ms delays
	Type(data []byte) error

	// Read reads from the stdout of the session.
	Read(data []byte) (int, error)

	// ReadRemaining reads the remaining bytes from stdout until EOF.
	ReadRemaining()

	// ReadRemainingStderr reads the remaining bytes from stderr until EOF.
	ReadRemainingStderr()

	// WaitForStdout waits for a specific byte sequence to appear on the stdout.
	WaitForStdout(ctx context.Context, data []byte) error

	// Stderr returns the reader for the stdout.
	Stderr() io.Reader

	// Wait waits for the session to terminate.
	Wait() error

	// ExitCode returns the exit code received from the session, or -1 if not received.
	ExitCode() int

	// Close closes the session.
	Close() error
}
