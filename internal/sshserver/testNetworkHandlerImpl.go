package sshserver

import (
	"context"
	"net"

    "go.containerssh.io/libcontainerssh/metadata"
)

type testNetworkHandlerImpl struct {
	AbstractNetworkConnectionHandler

	rootHandler  *testHandlerImpl
	client       net.TCPAddr
	connectionID string
	shutdown     bool
}

func (t *testNetworkHandlerImpl) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (SSHConnectionHandler, metadata.ConnectionAuthenticatedMetadata, error) {
	return &testSSHHandler{
		rootHandler:    t.rootHandler,
		networkHandler: t,
		metadata:       meta,
	}, meta, nil
}

func (t *testNetworkHandlerImpl) OnShutdown(_ context.Context) {
	t.shutdown = true
}
