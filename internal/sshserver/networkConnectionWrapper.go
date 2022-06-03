package sshserver

import (
	"context"

    "go.containerssh.io/libcontainerssh/metadata"
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
