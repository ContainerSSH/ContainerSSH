package sshserver

import (
	"context"

	"github.com/containerssh/libcontainerssh/metadata"
)

// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
type networkConnectionWrapper struct {
	NetworkConnectionHandler

	authenticatedMetadata metadata.ConnectionAuthenticatedMetadata
	sshConnectionHandler  SSHConnectionHandler
}

func (n *networkConnectionWrapper) OnShutdown(shutdownContext context.Context) {
	n.sshConnectionHandler.OnShutdown(shutdownContext)
}
