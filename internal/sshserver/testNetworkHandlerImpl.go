package sshserver

import (
	"context"
	"net"
)

type testNetworkHandlerImpl struct {
	AbstractNetworkConnectionHandler

	rootHandler  *testHandlerImpl
	client       net.TCPAddr
	connectionID string
	shutdown     bool
}

func (t *testNetworkHandlerImpl) OnHandshakeSuccess(username string, clientVersion string, metadata map[string]string) (
	connection SSHConnectionHandler,
	failureReason error,
) {
	return &testSSHHandler{
		rootHandler:    t.rootHandler,
		networkHandler: t,
		username:       username,
	}, nil
}

func (t *testNetworkHandlerImpl) OnShutdown(_ context.Context) {
	t.shutdown = true
}
