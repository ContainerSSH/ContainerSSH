package sshserver

import (
	"context"
	"net"
)

// testAuthenticationHandler is a conformanceTestHandler that authenticates and passes authentication to the configured backend.
type testAuthenticationHandler struct {
	users   []*TestUser
	backend Handler
}

func (t *testAuthenticationHandler) OnReady() error {
	return t.backend.OnReady()
}

func (t *testAuthenticationHandler) OnShutdown(ctx context.Context) {
	t.backend.OnShutdown(ctx)
}

func (t *testAuthenticationHandler) OnNetworkConnection(
	client net.TCPAddr,
	connectionID string,
) (NetworkConnectionHandler, error) {
	backend, err := t.backend.OnNetworkConnection(client, connectionID)
	if err != nil {
		return nil, err
	}

	return &testAuthenticationNetworkHandler{
		rootHandler: t,
		backend:     backend,
	}, nil
}
