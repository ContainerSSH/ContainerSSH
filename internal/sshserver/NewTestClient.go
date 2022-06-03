package sshserver

import (
    "go.containerssh.io/libcontainerssh/log"
	"golang.org/x/crypto/ssh"
)

// NewTestClient creates a new TestClient instance with the specified parameters
//
// - server is the host and IP pair of the server.
// - hostPrivateKey is the PEM-encoded private host key. The public key and fingerprint are automatically extracted.
// - username is the username.
// - password is the password used for authentication.
func NewTestClient(
	server string,
	hostPrivateKey string,
	user *TestUser,
	logger log.Logger,
) TestClient {
	private, err := ssh.ParsePrivateKey([]byte(hostPrivateKey))
	if err != nil {
		panic(err)
	}

	return &testClientImpl{
		server:  server,
		hostKey: private.PublicKey().Marshal(),
		user:    user,
		logger:  logger,
	}
}
