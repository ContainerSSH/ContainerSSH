package sshserver

// TestClientConnection is an individual established connection to the server
type TestClientConnection interface {
	// Session establishes a new session channel
	Session() (TestClientSession, error)

	//MustSession is identical to Session but panics if a session cannot be requested.
	MustSession() TestClientSession

	// Close closes the connection and all sessions in it.
	Close() error
}
