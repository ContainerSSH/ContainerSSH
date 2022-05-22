package sshserver

import (
	"context"
	"net"

	"github.com/containerssh/libcontainerssh/metadata"
)

// testHandlerImpl is a conformanceTestHandler implementation that fakes a "real" backend.
type testHandlerImpl struct {
	AbstractHandler

	shutdown bool
}

func (t *testHandlerImpl) OnShutdown(_ context.Context) {
	t.shutdown = true
}

func (t *testHandlerImpl) OnNetworkConnection(meta metadata.ConnectionMetadata) (NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return &testNetworkHandlerImpl{
		client:       net.TCPAddr(meta.RemoteAddress),
		connectionID: meta.ConnectionID,
		rootHandler:  t,
	}, meta, nil
}
