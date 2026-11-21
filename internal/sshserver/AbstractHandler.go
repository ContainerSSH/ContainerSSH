package sshserver

import (
	"context"
	"fmt"
	"net"
)

// AbstractHandler is the abstract implementation of the Handler interface that can be embedded to get a partial
// implementation.
type AbstractHandler struct {
}

// OnReady is called when the server is ready to receive connections. It has an opportunity to return an error to
//         abort the startup.
func (a *AbstractHandler) OnReady() error {
	return nil
}

// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
//            for the shutdown, after which the server should abort all running connections and return as fast as
//            possible.
func (a *AbstractHandler) OnShutdown(_ context.Context) {
}

// OnNetworkConnection is called when a new network connection is opened. It must either return a
// NetworkConnectionHandler object or an error. In case of an error the network connection is closed.
//
// The ip parameter provides the IP address of the connecting user. The connectionID parameter provides an opaque
// binary identifier for the connection that can be used to track the connection across multiple subsystems.
func (a *AbstractHandler) OnNetworkConnection(_ net.TCPAddr, _ string) (NetworkConnectionHandler, error) {
	return nil, fmt.Errorf("not implemented")
}
