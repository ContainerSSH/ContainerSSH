package sshserver

import (
	"bytes"
	"fmt"
	"net"

	"github.com/containerssh/libcontainerssh/log"
	messageCodes "github.com/containerssh/libcontainerssh/message"
	"golang.org/x/crypto/ssh"
)

type testClientImpl struct {
	server  string
	hostKey []byte
	user    *TestUser
	logger  log.Logger
}

func (t *testClientImpl) Connect() (TestClientConnection, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Connecting SSH server..."))
	sshConfig := &ssh.ClientConfig{
		User: t.user.Username(),
		Auth: t.user.GetAuthMethods(),
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if bytes.Equal(key.Marshal(), t.hostKey) {
			return nil
		}
		return fmt.Errorf("invalid host")
	}
	sshConnection, err := ssh.Dial("tcp", t.server, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("handshake failed (%w)", err)
	}

	return &testClientConnectionImpl{
		logger:        t.logger,
		sshConnection: sshConnection,
	}, nil
}

func (t *testClientImpl) MustConnect() TestClientConnection {
	connection, err := t.Connect()
	if err != nil {
		panic(err)
	}
	return connection
}
