package sshserver

import (
    "go.containerssh.io/libcontainerssh/log"
    messageCodes "go.containerssh.io/libcontainerssh/message"
	"golang.org/x/crypto/ssh"
)

type testClientConnectionImpl struct {
	sshConnection *ssh.Client
	logger        log.Logger
}

func (t *testClientConnectionImpl) MustSession() TestClientSession {
	session, err := t.Session()
	if err != nil {
		panic(err)
	}
	return session
}

func (t *testClientConnectionImpl) Session() (TestClientSession, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Opening session channel.."))
	session, err := t.sshConnection.NewSession()
	if err != nil {
		return nil, err
	}
	stdin := newSyncContextPipe()
	stdout := newSyncContextPipe()
	stderr := newSyncContextPipe()
	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	return &testClientSessionImpl{
		logger:   t.logger,
		session:  session,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		exitCode: -1,
	}, nil
}

func (t *testClientConnectionImpl) Close() error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Closing connection..."))
	return t.sshConnection.Close()
}
