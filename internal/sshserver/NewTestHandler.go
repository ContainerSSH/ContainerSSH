package sshserver

// NewTestHandler creates a conformanceTestHandler that can be used for testing purposes. It does not authenticate, that can be done
// using the NewTestAuthenticationHandler
func NewTestHandler() Handler {
	return &testHandlerImpl{}
}
