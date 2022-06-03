package sshserver

import (
	"context"

    "go.containerssh.io/libcontainerssh/metadata"
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
	meta metadata.ConnectionMetadata,
) (NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	backend, meta, err := t.backend.OnNetworkConnection(meta)
	if err != nil {
		return nil, meta, err
	}

	return &testAuthenticationNetworkHandler{
		rootHandler: t,
		backend:     backend,
	}, meta, nil
}
