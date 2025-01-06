package sshserver

// NewTestAuthenticationHandler creates a new backend that authenticates a user based on the users variable and passes
// all further calls to the backend.
func NewTestAuthenticationHandler(
	backend Handler,
	users ...*TestUser,
) Handler {
	return &testAuthenticationHandler{
		users:   users,
		backend: backend,
	}
}
