package sshserver

import (
	"context"
	"net"
)

// testHandlerImpl is a conformanceTestHandler implementation that fakes a "real" backend.
type testHandlerImpl struct {
	AbstractHandler

	shutdown bool
}

func (t *testHandlerImpl) OnShutdown(_ context.Context) {
	t.shutdown = true
}

func (t *testHandlerImpl) OnNetworkConnection(client net.TCPAddr, connectionID string) (NetworkConnectionHandler, error) {
	return &testNetworkHandlerImpl{
		client:       client,
		connectionID: connectionID,
		rootHandler:  t,
	}, nil
}
