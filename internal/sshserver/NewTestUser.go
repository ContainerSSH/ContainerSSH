package sshserver

// NewTestUser creates a user that can be used with NewTestHandler and NewTestClient.
func NewTestUser(username string) *TestUser {
	return &TestUser{
		username:            username,
		keyboardInteractive: map[string]string{},
	}
}
