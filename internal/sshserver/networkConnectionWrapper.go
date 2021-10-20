package sshserver

import (
	"context"
)

// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
type networkConnectionWrapper struct {
	NetworkConnectionHandler
	sshConnectionHandler SSHConnectionHandler
}

func (n *networkConnectionWrapper) OnShutdown(shutdownContext context.Context) {
	n.sshConnectionHandler.OnShutdown(shutdownContext)
}
